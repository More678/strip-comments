package textio

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func TestWriteExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.bak")
	raw := []byte("\xef\xbb\xbfraw\x00\r\nbytes\n")
	if err := WriteExclusive(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteExclusive(path, []byte("replacement"), 0o644); !os.IsExist(err) {
		t.Fatalf("existing backup must fail: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(raw) {
		t.Fatalf("backup changed: %q, %v", got, err)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" && st.Mode().Perm() != 0o600 {
		t.Fatalf("backup mode = %v", st.Mode())
	}
	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(entries) != 1 {
		t.Fatalf("temporary files left after conflict: %v, %v", entries, err)
	}
	if err := WriteExclusive(filepath.Join(path, "missing", "x"), raw, 0o600); err == nil {
		t.Fatal("invalid parent should fail")
	}
}

func TestWriteExclusiveRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "original")
	if err := os.WriteFile(target, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "linked.bak")
	if err := os.Symlink(target, path); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := WriteExclusive(path, []byte("replacement"), 0o600); !os.IsExist(err) {
		t.Fatalf("symlink backup must fail: %v", err)
	}
	got, err := os.ReadFile(target)
	if err != nil || string(got) != "keep me" {
		t.Fatalf("symlink target changed: %q, %v", got, err)
	}
}

func TestWriteExclusiveConcurrent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.bak")
	var wg sync.WaitGroup
	results := make(chan error, 8)
	for i := 0; i < cap(results); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- WriteExclusive(path, []byte("complete backup"), 0o600)
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		} else if !os.IsExist(err) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("exclusive publication succeeded %d times", successes)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "complete backup" {
		t.Fatalf("incomplete backup: %q, %v", got, err)
	}
}
