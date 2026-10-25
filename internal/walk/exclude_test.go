package walk

import (
	"path/filepath"
	"testing"
)

// TestExcludeBeatsNamedRoot pins the fix for -exclude being ignored on files
// and directories named on the command line: `-exclude main.go main.go` used
// to strip main.go anyway, which is the dangerous direction for a tool that
// rewrites files in place.
func TestExcludeBeatsNamedRoot(t *testing.T) {
	root := tree(t)

	t.Run("named file", func(t *testing.T) {
		files, _, errs := Walk(
			[]string{filepath.Join(root, "main.go")},
			Options{Exclude: []string{"main.go"}},
		)
		if len(errs) != 0 {
			t.Fatalf("errors: %v", errs)
		}
		if len(files) != 0 {
			t.Errorf("got %v, want nothing: -exclude must win", relPaths(t, root, files))
		}
	})

	t.Run("named dir", func(t *testing.T) {
		files, _, errs := Walk(
			[]string{filepath.Join(root, "sub")},
			Options{Exclude: []string{"sub"}},
		)
		if len(errs) != 0 {
			t.Fatalf("errors: %v", errs)
		}
		if len(files) != 0 {
			t.Errorf("got %v, want nothing", relPaths(t, root, files))
		}
	})

	t.Run("other roots still processed", func(t *testing.T) {
		files, _, errs := Walk(
			[]string{filepath.Join(root, "main.go"), filepath.Join(root, "util.go")},
			Options{Exclude: []string{"main.go"}},
		)
		if len(errs) != 0 {
			t.Fatalf("errors: %v", errs)
		}
		if got := relPaths(t, root, files); !equalPaths(got, []string{"util.go"}) {
			t.Errorf("got %v, want [util.go]", got)
		}
	})
}

// TestExcludePathGlobAbsoluteRoot pins the fix for path globs only working
// under a relative root: `sub/*.go` is matched against the path relative to
// the root as well, so an absolute root behaves the same as ".".
func TestExcludePathGlobAbsoluteRoot(t *testing.T) {
	root := tree(t)
	abs, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}

	files, _, errs := Walk([]string{abs}, Options{
		Exclude: []string{"sub/*.go"},
		Exts:    []string{".go"},
	})
	if len(errs) != 0 {
		t.Fatalf("errors: %v", errs)
	}
	got := relPaths(t, root, files)
	for _, p := range got {
		if p == "sub/deep.go" {
			t.Fatalf("sub/deep.go survived %q under an absolute root: got %v",
				"sub/*.go", got)
		}
	}
	// The glob must not swallow .go files outside sub/.
	if !equalPaths(got, []string{"main.go", "util.go"}) {
		t.Errorf("got %v, want [main.go util.go]", got)
	}
}
