// Command strip-comments removes every comment from the source files under
// the given paths (the current folder by default).
//
// Nothing is written unless -w is passed: a plain run only reports what it
// would change.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/More678/Remove-All-Commet/internal/lang"
	"github.com/More678/Remove-All-Commet/internal/scan"
	"github.com/More678/Remove-All-Commet/internal/textio"
	"github.com/More678/Remove-All-Commet/internal/walk"
)

// version is overridden at build time:
// go build -ldflags "-X main.version=v1.2.3"
var version = "dev"

// Exit codes.
const (
	exitOK    = 0
	exitError = 1
	exitUsage = 2
)

type config struct {
	write   bool
	backup  bool
	stdout  bool
	force   bool
	risky   bool
	quiet   bool
	lang    *lang.Language
	workers int
}

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	fs := flag.NewFlagSet("strip-comments", flag.ExitOnError)
	fs.Usage = func() {
		out := fs.Output()
		fmt.Fprintf(out, "usage: strip-comments [flags] [path ...]\n\n")
		fmt.Fprintf(out, "Removes all comments from the source files under each path\n")
		fmt.Fprintf(out, "(default \".\"). Without -w nothing is written: the run only\n")
		fmt.Fprintf(out, "reports the files it would change.\n\n")
		fs.PrintDefaults()
	}

	var (
		write     = fs.Bool("w", false, "rewrite the files in place (default: dry run)")
		writeLong = fs.Bool("write", false, "alias for -w")
		backup    = fs.Bool("backup", false, "keep <file>.bak before rewriting (with -w)")
		toStdout  = fs.Bool("stdout", false, "print one stripped file to stdout, write nothing")
		langName  = fs.String("lang", "", "force this language (see -list)")
		extList   = fs.String("ext", "", "only these extensions, e.g. \".go,.rs\"")
		exclude   = fs.String("exclude", "", "extra names or globs to skip, comma separated")
		keepDoc   = fs.Bool("keep-doc", false, "keep documentation comments (///, /** */)")
		keepBlank = fs.Bool("keep-blank", false, "leave an empty line instead of deleting it")
		keepLic   = fs.Bool("keep-license", false, "keep the leading file comment")
		force     = fs.Bool("force", false, "strip even when the file looks unsafe")
		risky     = fs.Bool("risky", false, "include languages with heuristic support")
		jobs      = fs.Int("j", runtime.NumCPU(), "number of parallel workers")
		quiet     = fs.Bool("q", false, "only report errors")
		list      = fs.Bool("list", false, "list the supported languages and exit")
		showVer   = fs.Bool("version", false, "print the version and exit")
		maxSize   = fs.Int64("max-size", walk.DefaultMaxSize, "skip files larger than this many bytes")
	)
	if err := fs.Parse(args); err != nil {
		return exitUsage
	}

	switch {
	case *showVer:
		fmt.Printf("strip-comments %s\n", version)
		return exitOK
	case *list:
		printLanguages()
		return exitOK
	}

	cfg := config{
		write:   *write || *writeLong,
		backup:  *backup,
		stdout:  *toStdout,
		force:   *force,
		risky:   *risky,
		quiet:   *quiet,
		workers: *jobs,
	}
	if cfg.workers < 1 {
		cfg.workers = 1
	}
	if *langName != "" {
		cfg.lang = lang.ByName(*langName)
		if cfg.lang == nil {
			fmt.Fprintf(os.Stderr, "unknown language %q, try -list\n", *langName)
			return exitUsage
		}
	}
	opt := scan.Options{KeepDoc: *keepDoc, KeepBlank: *keepBlank,
		KeepLicense: *keepLic, Force: *force}

	roots := fs.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}
	if cfg.stdout {
		if len(roots) != 1 {
			fmt.Fprintln(os.Stderr, "-stdout needs exactly one file")
			return exitUsage
		}
		return runStdout(roots[0], cfg, opt)
	}

	self, _ := os.Executable()
	files, skips, errs := walk.Walk(roots, walk.Options{
		Lang:     cfg.lang,
		Exts:     splitList(*extList),
		Exclude:  splitList(*exclude),
		Risky:    cfg.risky,
		MaxSize:  *maxSize,
		SelfPath: self,
	})
	for _, err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
	results := process(files, cfg, opt)
	return report(results, skips, len(errs), cfg)
}

// result is what happened to one file.
type result struct {
	file     walk.File
	changed  bool
	comments int
	lines    int
	skip     string
	err      error
}

// process strips every file through a worker pool. Results stay in walk
// order, so the report is deterministic whatever the workers do.
func process(files []walk.File, cfg config, opt scan.Options) []result {
	results := make([]result, len(files))
	if len(files) == 0 {
		return results
	}
	workers := cfg.workers
	if workers > len(files) {
		workers = len(files)
	}
	jobs := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				results[idx] = stripOne(files[idx], cfg, opt)
			}
		}()
	}
	for i := range files {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return results
}

// stripOne reads, strips and (in write mode) rewrites a single file.
func stripOne(f walk.File, cfg config, opt scan.Options) result {
	res := result{file: f}
	raw, err := os.ReadFile(f.Path)
	if err != nil {
		res.err = err
		return res
	}
	if textio.IsBinary(raw) {
		res.skip = "binary file"
		return res
	}
	text, format := textio.Decode(raw)
	sr := scan.Strip(text, f.Lang, opt)
	if sr.Unsafe && !cfg.force {
		res.skip = sr.Reason
		return res
	}
	if sr.Comments == 0 {
		// Nothing was stripped, so leave the bytes alone. Re-encoding would
		// still rewrite a file whose line endings are mixed, and normalising
		// line endings is not this tool's job.
		return res
	}
	out := textio.Encode(sr.Output, format)
	if bytes.Equal(out, raw) {
		return res
	}
	res.changed = true
	res.comments = sr.Comments
	res.lines = sr.RemovedLines
	if !cfg.write {
		return res
	}
	if cfg.backup {
		if err := textio.WriteAtomicMode(f.Path+".bak", raw, f.Mode.Perm()); err != nil {
			res.err = err
			return res
		}
	}
	if err := textio.WriteAtomic(f.Path, out); err != nil {
		res.err = err
	}
	return res
}

// report prints one line per changed file, then the summary, and returns the
// exit code.
func report(results []result, skips []walk.Skip, walkErrs int, cfg config) int {
	var changed, comments, lines, errCount int
	skipped := len(skips)

	for _, r := range results {
		switch {
		case r.err != nil:
			errCount++
			fmt.Fprintf(os.Stderr, "%s: %v\n", r.file.Path, r.err)
		case r.skip != "":
			skipped++
			if !cfg.quiet {
				fmt.Printf("SKIP %s: %s\n", displayPath(r.file.Path), r.skip)
			}
		case r.changed:
			changed++
			comments += r.comments
			lines += r.lines
			if !cfg.quiet {
				fmt.Printf("  %s  %s  -%d comments, -%d lines\n",
					displayPath(r.file.Path), r.file.Lang.Name, r.comments, r.lines)
			}
		}
	}
	if !cfg.quiet {
		for _, s := range skips {
			fmt.Printf("SKIP %s: %s\n", displayPath(s.Path), s.Reason)
		}
		fmt.Printf("\n%d files scanned, %d changed, %d comments removed, %d lines removed, %d skipped, %d errors\n",
			len(results), changed, comments, lines, skipped, errCount+walkErrs)
		if !cfg.write && changed > 0 {
			fmt.Println("dry run: re-run with -w to apply")
		}
	}
	if errCount+walkErrs > 0 {
		return exitError
	}
	return exitOK
}

// runStdout strips one file and writes the result to stdout.
func runStdout(path string, cfg config, opt scan.Options) int {
	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	if textio.IsBinary(raw) {
		fmt.Fprintf(os.Stderr, "%s: binary file\n", path)
		return exitError
	}
	l := cfg.lang
	if l == nil {
		l = lang.ByFilename(filepath.Base(path))
	}
	if l == nil {
		fmt.Fprintf(os.Stderr, "%s: unknown language, use -lang\n", path)
		return exitError
	}
	text, format := textio.Decode(raw)
	sr := scan.Strip(text, l, opt)
	if sr.Unsafe && !cfg.force {
		fmt.Fprintf(os.Stderr, "%s: %s (use -force)\n", path, sr.Reason)
		return exitError
	}
	if _, err := os.Stdout.Write(textio.Encode(sr.Output, format)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitError
	}
	return exitOK
}

// printLanguages lists every supported language.
func printLanguages() {
	for _, l := range lang.All() {
		risky := ""
		if l.Risky {
			risky = "  (risky)"
		}
		names := append(append([]string{}, l.Exts...), l.Filenames...)
		fmt.Printf("%-12s %s%s\n", l.Name, strings.Join(names, " "), risky)
	}
}

// displayPath prefers a path relative to the working directory.
func displayPath(path string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, path); err == nil && !strings.HasPrefix(rel, "..") {
			return rel
		}
	}
	return path
}

// splitList parses a comma separated flag value.
func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if part = strings.TrimSpace(part); part != "" {
			out = append(out, part)
		}
	}
	return out
}
