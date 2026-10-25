package walk

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/More678/strip-comments/internal/lang"
)

// write creates path (with its parents) under dir and returns nothing.
func write(t *testing.T, dir, path, body string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// tree builds the fixture used by most tests and returns its root.
func tree(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, f := range []string{
		"main.go", "util.go", "app.py", "lib.c", "style.css",
		"sub/deep.go", "sub/notes.md", "sub/Makefile",
		"docker/Dockerfile", "boot.asm",
		".git/config.go", "node_modules/pkg/index.js", "vendor/dep/dep.go",
		"__pycache__/cached.py", "dist/bundle.js",
		"keep.go.bak", "sub/old.BAK",
	} {
		write(t, root, f, "x = 1 # c\n")
	}
	return root
}

// relPaths turns the result into root-relative slash paths for comparison.
func relPaths(t *testing.T, root string, files []File) []string {
	t.Helper()
	out := make([]string, 0, len(files))
	for _, f := range files {
		rel, err := filepath.Rel(root, f.Path)
		if err != nil {
			t.Fatalf("Rel(%q, %q): %v", root, f.Path, err)
		}
		out = append(out, filepath.ToSlash(rel))
	}
	return out
}

func equalPaths(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestWalkDefault covers the whole default policy at once: skip dirs are
// pruned, .bak is invisible, risky files are skipped with a reason, and the
// result is sorted.
func TestWalkDefault(t *testing.T) {
	root := tree(t)
	files, skips, errs := Walk([]string{root}, Options{})
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	want := []string{
		"app.py", "docker/Dockerfile", "lib.c", "main.go", "style.css",
		"sub/Makefile", "sub/deep.go", "util.go",
	}
	got := relPaths(t, root, files)
	if !equalPaths(got, want) {
		t.Errorf("files =\n  %v\nwant\n  %v", got, want)
	}

	// boot.asm is risky, so it is reported rather than silently dropped.
	if len(skips) != 1 {
		t.Fatalf("skips = %v, want just boot.asm", skips)
	}
	if filepath.Base(skips[0].Path) != "boot.asm" {
		t.Errorf("skip path = %q, want boot.asm", skips[0].Path)
	}
	if !strings.Contains(skips[0].Reason, "risky") {
		t.Errorf("skip reason = %q, want it to mention risky", skips[0].Reason)
	}
}

func TestWalkPrunesSkipDirs(t *testing.T) {
	root := tree(t)
	// This is an absence check, so assert the bait exists first: otherwise a
	// fixture that stopped creating these files would pass for the wrong
	// reason.
	bait := map[string]string{
		".git/":         ".git/config.go",
		"node_modules/": "node_modules/pkg/index.js",
		"vendor/":       "vendor/dep/dep.go",
		"__pycache__/":  "__pycache__/cached.py",
		"dist/":         "dist/bundle.js",
	}
	for _, rel := range bait {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			t.Fatalf("fixture is missing %s: %v", rel, err)
		}
	}
	files, _, _ := Walk([]string{root}, Options{})
	for _, p := range relPaths(t, root, files) {
		for prefix := range bait {
			if strings.HasPrefix(p, prefix) {
				t.Errorf("%q was not pruned", p)
			}
		}
	}
	// Every skip dir in the exported list is matched case-insensitively.
	for _, d := range []string{"NODE_MODULES", "Vendor", ".GIT"} {
		if !isSkipDir(d) {
			t.Errorf("isSkipDir(%q) = false, want true", d)
		}
	}
	if isSkipDir("src") {
		t.Error(`isSkipDir("src") = true, want false`)
	}
}

func TestWalkIgnoresBakFiles(t *testing.T) {
	root := tree(t)
	files, skips, _ := Walk([]string{root}, Options{})
	for _, p := range relPaths(t, root, files) {
		if strings.EqualFold(filepath.Ext(p), ".bak") {
			t.Errorf("%q must never be returned", p)
		}
	}
	// .bak is dropped silently, not reported as a skip.
	for _, s := range skips {
		if strings.EqualFold(filepath.Ext(s.Path), ".bak") {
			t.Errorf("%q must not be reported as a skip", s.Path)
		}
	}
	// Naming one explicitly does not help either.
	files, _, _ = Walk([]string{filepath.Join(root, "keep.go.bak")}, Options{})
	if len(files) != 0 {
		t.Errorf("explicit .bak returned %v, want none", relPaths(t, root, files))
	}
}

func TestWalkSorted(t *testing.T) {
	root := tree(t)
	// Roots in reverse order must still come back sorted.
	files, _, _ := Walk([]string{
		filepath.Join(root, "util.go"),
		filepath.Join(root, "sub"),
		filepath.Join(root, "app.py"),
	}, Options{})
	got := relPaths(t, root, files)
	want := []string{"app.py", "sub/Makefile", "sub/deep.go", "util.go"}
	if !equalPaths(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
}

// TestWalkExplicitSkipDir: SkipDirs only applies below a root, so naming a
// skipped directory on the command line descends into it.
func TestWalkExplicitSkipDir(t *testing.T) {
	root := tree(t)
	files, _, _ := Walk([]string{filepath.Join(root, "vendor")}, Options{})
	got := relPaths(t, root, files)
	if want := []string{"vendor/dep/dep.go"}; !equalPaths(got, want) {
		t.Errorf("explicit vendor root = %v, want %v", got, want)
	}

	// The same directory under a walked root stays pruned.
	files, _, _ = Walk([]string{root}, Options{})
	for _, p := range relPaths(t, root, files) {
		if strings.HasPrefix(p, "vendor/") {
			t.Errorf("%q should be pruned when reached from a parent root", p)
		}
	}
}

func TestWalkExclude(t *testing.T) {
	root := tree(t)
	cases := []struct {
		name    string
		exclude []string
		gone    []string
		kept    []string
	}{
		{
			name: "exact name", exclude: []string{"util.go"},
			gone: []string{"util.go"}, kept: []string{"main.go"},
		},
		{
			name: "glob", exclude: []string{"*.go"},
			gone: []string{"main.go", "util.go", "sub/deep.go"},
			kept: []string{"app.py", "lib.c"},
		},
		{
			name: "directory name", exclude: []string{"sub"},
			gone: []string{"sub/deep.go", "sub/Makefile"},
			kept: []string{"main.go"},
		},
		{
			// Matching is case-insensitive on the base name.
			name: "case insensitive", exclude: []string{"MAIN.GO"},
			gone: []string{"main.go"}, kept: []string{"util.go"},
		},
		{
			name: "several patterns", exclude: []string{"*.py", "*.css", "sub"},
			gone: []string{"app.py", "style.css", "sub/deep.go"},
			kept: []string{"main.go", "lib.c"},
		},
		{
			// Blank entries are ignored rather than matching everything.
			name: "blank pattern", exclude: []string{"", "   "},
			kept: []string{"main.go", "app.py"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files, _, _ := Walk([]string{root}, Options{Exclude: tc.exclude})
			got := relPaths(t, root, files)
			has := map[string]bool{}
			for _, p := range got {
				has[p] = true
			}
			for _, p := range tc.gone {
				if has[p] {
					t.Errorf("%q should have been excluded (got %v)", p, got)
				}
			}
			for _, p := range tc.kept {
				if !has[p] {
					t.Errorf("%q should have been kept (got %v)", p, got)
				}
			}
		})
	}
}

// TestWalkExcludePathGlob covers a pattern containing a separator, which is
// matched against the whole path instead of the base name. The root has to be
// relative for the pattern to line up, so this one runs in the tree.
func TestWalkExcludePathGlob(t *testing.T) {
	root := tree(t)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(cwd) })

	files, _, _ := Walk([]string{"."}, Options{Exclude: []string{"sub/*.go"}})
	var got []string
	for _, f := range files {
		got = append(got, filepath.ToSlash(filepath.Clean(f.Path)))
	}
	has := map[string]bool{}
	for _, p := range got {
		has[p] = true
	}
	if has["sub/deep.go"] {
		t.Errorf("sub/deep.go should match the path glob (got %v)", got)
	}
	if !has["main.go"] {
		t.Errorf("main.go should be kept (got %v)", got)
	}
	// The same glob must not knock out a same-named file elsewhere.
	if !has["sub/Makefile"] {
		t.Errorf("sub/Makefile should be kept (got %v)", got)
	}
}

// TestWalkLanguageResolution checks both lookup routes: by extension and by
// exact base name for files that have none.
func TestWalkLanguageResolution(t *testing.T) {
	root := tree(t)
	files, _, _ := Walk([]string{root}, Options{})
	byRel := map[string]*lang.Language{}
	for i, p := range relPaths(t, root, files) {
		byRel[p] = files[i].Lang
	}
	for rel, want := range map[string]string{
		"main.go":           "go",
		"app.py":            "python",
		"lib.c":             "c",
		"style.css":         "css",
		"sub/Makefile":      "makefile",
		"docker/Dockerfile": "dockerfile",
	} {
		l, ok := byRel[rel]
		if !ok {
			t.Errorf("%s is missing from the walk", rel)
			continue
		}
		if l == nil {
			t.Errorf("%s: Lang is nil", rel)
			continue
		}
		if l.Name != want {
			t.Errorf("%s: Lang = %q, want %q", rel, l.Name, want)
		}
	}
	// An unknown extension is dropped silently during a directory walk.
	if _, ok := byRel["sub/notes.md"]; ok {
		t.Error("sub/notes.md has no language and must not be returned")
	}
}

// TestWalkUnknownExplicit: a file named on the command line reports why it
// was skipped instead of vanishing.
func TestWalkUnknownExplicit(t *testing.T) {
	root := tree(t)
	files, skips, _ := Walk([]string{filepath.Join(root, "sub/notes.md")}, Options{})
	if len(files) != 0 {
		t.Errorf("files = %v, want none", relPaths(t, root, files))
	}
	if len(skips) != 1 {
		t.Fatalf("skips = %v, want one", skips)
	}
	if !strings.Contains(skips[0].Reason, "unknown language") {
		t.Errorf("reason = %q, want it to mention the unknown language", skips[0].Reason)
	}
}

func TestWalkRisky(t *testing.T) {
	root := tree(t)
	asm := filepath.Join(root, "boot.asm")

	files, skips, _ := Walk([]string{root}, Options{})
	for _, p := range relPaths(t, root, files) {
		if p == "boot.asm" {
			t.Error("boot.asm is risky and must be skipped by default")
		}
	}
	if len(skips) == 0 {
		t.Error("the risky file should be reported as a skip")
	}

	files, _, _ = Walk([]string{root}, Options{Risky: true})
	found := false
	for _, p := range relPaths(t, root, files) {
		if p == "boot.asm" {
			found = true
		}
	}
	if !found {
		t.Error("boot.asm should be included with Risky")
	}

	// Naming it explicitly is not enough: the guard still applies.
	files, skips, _ = Walk([]string{asm}, Options{})
	if len(files) != 0 {
		t.Errorf("explicit risky file = %v, want none without Risky", relPaths(t, root, files))
	}
	if len(skips) != 1 || !strings.Contains(skips[0].Reason, "risky") {
		t.Errorf("skips = %v, want a risky reason", skips)
	}

	// But -ext counts as an explicit choice and lifts the guard.
	files, _, _ = Walk([]string{root}, Options{Exts: []string{".asm"}})
	if got := relPaths(t, root, files); !equalPaths(got, []string{"boot.asm"}) {
		t.Errorf("-ext .asm = %v, want [boot.asm] (Exts lifts the Risky guard)", got)
	}
}

func TestWalkExtNarrowing(t *testing.T) {
	root := tree(t)
	cases := []struct {
		name string
		exts []string
		want []string
	}{
		{"one ext", []string{".go"}, []string{"main.go", "sub/deep.go", "util.go"}},
		{"two exts", []string{".py", ".c"}, []string{"app.py", "lib.c"}},
		// normExt adds the dot and lower-cases, so these are all the same.
		{"no dot", []string{"go"}, []string{"main.go", "sub/deep.go", "util.go"}},
		{"upper case", []string{".GO"}, []string{"main.go", "sub/deep.go", "util.go"}},
		{"padded", []string{"  .go  "}, []string{"main.go", "sub/deep.go", "util.go"}},
		{"unknown ext", []string{".nope"}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			files, _, _ := Walk([]string{root}, Options{Exts: tc.exts})
			got := relPaths(t, root, files)
			if !equalPaths(got, tc.want) {
				t.Errorf("files = %v, want %v", got, tc.want)
			}
		})
	}
	// Narrowing by extension excludes files that have none, even known ones.
	files, _, _ := Walk([]string{root}, Options{Exts: []string{".go"}})
	for _, p := range relPaths(t, root, files) {
		if strings.HasSuffix(p, "Makefile") || strings.HasSuffix(p, "Dockerfile") {
			t.Errorf("%q has no extension and must not match -ext .go", p)
		}
	}
}

func TestWalkMaxSize(t *testing.T) {
	root := t.TempDir()
	write(t, root, "small.go", "// c\n")
	write(t, root, "big.go", strings.Repeat("// c\n", 100))
	bigSize := int64(len("// c\n") * 100)

	files, skips, _ := Walk([]string{root}, Options{MaxSize: bigSize - 1})
	if got := relPaths(t, root, files); !equalPaths(got, []string{"small.go"}) {
		t.Errorf("files = %v, want [small.go]", got)
	}
	if len(skips) != 1 || !strings.Contains(skips[0].Reason, "max size") {
		t.Fatalf("skips = %v, want big.go over max size", skips)
	}
	if filepath.Base(skips[0].Path) != "big.go" {
		t.Errorf("skip = %q, want big.go", skips[0].Path)
	}

	// Exactly at the limit is kept: the check is size > MaxSize.
	files, _, _ = Walk([]string{root}, Options{MaxSize: bigSize})
	if got := relPaths(t, root, files); !equalPaths(got, []string{"big.go", "small.go"}) {
		t.Errorf("at the limit files = %v, want both", got)
	}

	// MaxSize <= 0 falls back to DefaultMaxSize rather than skipping all.
	files, _, _ = Walk([]string{root}, Options{MaxSize: 0})
	if len(files) != 2 {
		t.Errorf("MaxSize 0 returned %d files, want 2 via DefaultMaxSize", len(files))
	}
}

func TestWalkDeduplicates(t *testing.T) {
	root := tree(t)
	main := filepath.Join(root, "main.go")
	files, _, _ := Walk([]string{main, main, root}, Options{Exts: []string{".go"}})
	count := 0
	for _, p := range relPaths(t, root, files) {
		if p == "main.go" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("main.go appears %d times, want 1", count)
	}
}

func TestWalkSelfPath(t *testing.T) {
	root := tree(t)
	self := filepath.Join(root, "main.go")
	files, _, _ := Walk([]string{root}, Options{SelfPath: self})
	for _, p := range relPaths(t, root, files) {
		if p == "main.go" {
			t.Error("the tool's own path must never be returned")
		}
	}
}

func TestWalkForcedLang(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.go", "// c\n")
	write(t, root, "weird.xyz", "// c\n")
	goL := lang.ByName("go")
	if goL == nil {
		t.Skip("go is not registered")
	}

	// In a directory walk without -ext, -lang only picks files that already
	// resolve to that language, so an unknown extension stays untouched.
	files, _, _ := Walk([]string{root}, Options{Lang: goL})
	if got := relPaths(t, root, files); !equalPaths(got, []string{"a.go"}) {
		t.Errorf("files = %v, want [a.go]", got)
	}

	// Named explicitly, the forced language applies.
	files, _, _ = Walk([]string{filepath.Join(root, "weird.xyz")}, Options{Lang: goL})
	if len(files) != 1 || files[0].Lang != goL {
		t.Fatalf("explicit file with -lang = %v, want one go file", relPaths(t, root, files))
	}

	// With -ext the forced language applies to the selected files too.
	files, _, _ = Walk([]string{root}, Options{Lang: goL, Exts: []string{".xyz"}})
	if got := relPaths(t, root, files); !equalPaths(got, []string{"weird.xyz"}) {
		t.Fatalf("files = %v, want [weird.xyz]", got)
	}
	if files[0].Lang != goL {
		t.Errorf("Lang = %q, want go", files[0].Lang.Name)
	}
}

func TestWalkFileMetadata(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.go", "// c\n")
	files, _, _ := Walk([]string{root}, Options{})
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}
	if files[0].Size != int64(len("// c\n")) {
		t.Errorf("Size = %d, want %d", files[0].Size, len("// c\n"))
	}
	if !files[0].Mode.IsRegular() {
		t.Errorf("Mode = %v, want a regular file", files[0].Mode)
	}
}

func TestWalkMissingRoot(t *testing.T) {
	root := t.TempDir()
	files, _, errs := Walk([]string{filepath.Join(root, "nope")}, Options{})
	if len(errs) != 1 {
		t.Errorf("errs = %v, want one for the missing root", errs)
	}
	if len(files) != 0 {
		t.Errorf("files = %d, want 0", len(files))
	}
}
