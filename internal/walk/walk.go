// Package walk collects the files to process from a set of roots and decides
// which language each one is written in.
package walk

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/More678/Remove-All-Commet/internal/lang"
)

// SkipDirs are never descended into. A directory named on the command line is
// still processed: the list only applies below a root.
var SkipDirs = []string{
	".git", ".hg", ".svn", "node_modules", "vendor", "target", "build",
	"dist", "out", "bin", "obj", "__pycache__", ".venv", "venv", "env",
	".idea", ".vscode", ".tox", ".mypy_cache", ".pytest_cache", ".gradle",
	".next", ".cache",
}

// DefaultMaxSize is the largest file considered, 10 MiB.
const DefaultMaxSize int64 = 10 << 20

// Options configures a walk.
type Options struct {
	Lang     *lang.Language // force this language for every file
	Exts     []string       // restrict to these extensions (".go"), empty = all
	Exclude  []string       // extra names or globs to skip
	Risky    bool           // include languages flagged Risky
	MaxSize  int64          // skip larger files; 0 means DefaultMaxSize
	SelfPath string         // the tool's own binary, never touched
}

// File is one file worth stripping.
type File struct {
	Path string
	Lang *lang.Language
	Size int64
	Mode os.FileMode
}

// Skip is a file that was found but deliberately left alone.
type Skip struct {
	Path   string
	Reason string
}

type walker struct {
	opt     Options
	exts    map[string]bool
	files   []File
	skips   []Skip
	errs    []error
	seen    map[string]bool
	curRoot string // the root currently being walked, for relative glob matching
}

// Walk expands roots into a deterministic, de-duplicated list of files.
// Directories are traversed recursively; symlinks below a root are ignored.
func Walk(roots []string, opt Options) ([]File, []Skip, []error) {
	if opt.MaxSize <= 0 {
		opt.MaxSize = DefaultMaxSize
	}
	w := &walker{opt: opt, seen: map[string]bool{}}
	if len(opt.Exts) > 0 {
		w.exts = map[string]bool{}
		for _, e := range opt.Exts {
			w.exts[normExt(e)] = true
		}
	}
	if len(roots) == 0 {
		roots = []string{"."}
	}
	for _, root := range roots {
		clean := filepath.Clean(root)
		w.curRoot = clean
		w.root(clean)
	}
	sort.Slice(w.files, func(i, j int) bool { return w.files[i].Path < w.files[j].Path })
	sort.Slice(w.skips, func(i, j int) bool { return w.skips[i].Path < w.skips[j].Path })
	return w.files, w.skips, w.errs
}

// root handles one command line argument. Stat follows a symlinked root, so
// naming a linked directory explicitly still works.
func (w *walker) root(path string) {
	// -exclude wins over a named root: `-exclude main.go main.go` must not
	// strip main.go. Silently dropping it is safer than writing to a file the
	// user asked to leave alone.
	if w.excluded(filepath.Base(path), path) {
		return
	}
	st, err := os.Stat(path)
	if err != nil {
		w.errs = append(w.errs, err)
		return
	}
	if st.IsDir() {
		w.dir(path)
		return
	}
	if !st.Mode().IsRegular() {
		w.skips = append(w.skips, Skip{path, "not a regular file"})
		return
	}
	w.file(path, st.Size(), st.Mode(), true)
}

func (w *walker) dir(dir string) {
	ents, err := os.ReadDir(dir) // already sorted by name
	if err != nil {
		w.errs = append(w.errs, err)
		return
	}
	for _, e := range ents {
		name := e.Name()
		path := filepath.Join(dir, name)
		if w.excluded(name, path) {
			continue
		}
		if e.IsDir() {
			if isSkipDir(name) {
				continue
			}
			w.dir(path)
			continue
		}
		if e.Type()&os.ModeSymlink != 0 {
			continue // never rewrite through a link
		}
		if !e.Type().IsRegular() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			w.errs = append(w.errs, err)
			continue
		}
		w.file(path, info.Size(), info.Mode(), false)
	}
}

// file applies the size, language and risk rules. explicit marks a file named
// on the command line, which reports why it was skipped instead of staying
// silent.
func (w *walker) file(path string, size int64, mode os.FileMode, explicit bool) {
	base := filepath.Base(path)
	if strings.EqualFold(filepath.Ext(base), ".bak") {
		return
	}
	if w.isSelf(path) {
		return
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	key := strings.ToLower(filepath.Clean(abs))
	if w.seen[key] {
		return
	}

	l, forced := w.language(base, explicit)
	if l == nil {
		if explicit {
			w.seen[key] = true
			w.skips = append(w.skips, Skip{path, "unknown language, use -lang"})
		}
		return
	}
	if l.Risky && !w.opt.Risky && !forced {
		w.seen[key] = true
		w.skips = append(w.skips, Skip{path, "risky language " + l.Name + ", use -risky"})
		return
	}
	if size > w.opt.MaxSize {
		w.seen[key] = true
		w.skips = append(w.skips, Skip{path, "larger than max size"})
		return
	}
	w.seen[key] = true
	w.files = append(w.files, File{Path: path, Lang: l, Size: size, Mode: mode})
}

// language resolves the language for base. forced reports an explicit choice
// through -lang or -ext, which also lifts the Risky guard.
//
// A forced -lang applies to files named on the command line and to files
// selected by -ext. Inside a directory walk without -ext it only picks files
// that already resolve to that language, so `-lang go .` does not try to
// parse README.md or images as Go.
func (w *walker) language(base string, explicit bool) (l *lang.Language, forced bool) {
	ext := normExt(filepath.Ext(base))
	if w.exts != nil && !w.exts[ext] {
		return nil, false
	}
	if w.opt.Lang != nil {
		if w.exts == nil && !explicit && lang.ByFilename(base) != w.opt.Lang {
			return nil, false
		}
		return w.opt.Lang, true
	}
	return lang.ByFilename(base), w.exts != nil
}

func (w *walker) isSelf(path string) bool {
	if w.opt.SelfPath == "" {
		return false
	}
	a, err1 := filepath.Abs(path)
	b, err2 := filepath.Abs(w.opt.SelfPath)
	if err1 != nil || err2 != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

// excluded matches a user pattern against the base name, and against the
// whole path when the pattern contains a separator. A path pattern is tried
// against the path relative to the current root as well, so `sub/*.go` works
// whether the root is "." or an absolute directory.
func (w *walker) excluded(name, path string) bool {
	lname := strings.ToLower(name)
	for _, pat := range w.opt.Exclude {
		pat = strings.ToLower(strings.TrimSpace(pat))
		if pat == "" {
			continue
		}
		if pat == lname {
			return true
		}
		if ok, err := filepath.Match(pat, lname); err == nil && ok {
			return true
		}
		if !strings.ContainsAny(pat, `/\`) {
			continue
		}
		pat = filepath.ToSlash(pat)
		for _, cand := range w.pathForms(path) {
			if ok, err := filepath.Match(pat, cand); err == nil && ok {
				return true
			}
		}
	}
	return false
}

// pathForms lists the slash-separated, lower-cased spellings of path that a
// user-supplied glob may reasonably be written against.
func (w *walker) pathForms(path string) []string {
	forms := []string{strings.ToLower(filepath.ToSlash(path))}
	if w.curRoot != "" && w.curRoot != "." {
		if rel, err := filepath.Rel(w.curRoot, path); err == nil &&
			!strings.HasPrefix(rel, "..") {
			forms = append(forms, strings.ToLower(filepath.ToSlash(rel)))
		}
	}
	return forms
}

func isSkipDir(name string) bool {
	for _, d := range SkipDirs {
		if strings.EqualFold(name, d) {
			return true
		}
	}
	return false
}

// normExt lower-cases an extension and makes sure it has a leading dot, so
// "-ext go" and "-ext .GO" behave the same.
func normExt(e string) string {
	e = strings.ToLower(strings.TrimSpace(e))
	if e != "" && !strings.HasPrefix(e, ".") {
		e = "." + e
	}
	return e
}
