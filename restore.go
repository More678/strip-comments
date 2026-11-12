package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/More678/strip-comments/internal/textio"
	"github.com/More678/strip-comments/internal/walk"
)

// runRestore deliberately requires named files: no recursive rollback by accident.
func runRestore(paths []string, write, quiet, allowSecrets bool) int {
	code := exitOK
	for _, path := range paths {
		changed, err := restoreOne(path, write, allowSecrets)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			code = exitError
			continue
		}
		if quiet || !changed {
			continue
		}
		if write {
			fmt.Printf("restored %s; previous contents saved in %s.restore.bak\n", path, path)
		} else {
			fmt.Printf("would restore %s from %s.bak (use -restore -w)\n", path, path)
		}
	}
	return code
}

func restoreOne(path string, write, allowSecrets bool) (bool, error) {
	if strings.EqualFold(filepath.Ext(path), ".bak") {
		return false, fmt.Errorf("name the source file, not its .bak backup")
	}
	if !allowSecrets {
		if why := walk.SecretReason(path); why != "" {
			return false, fmt.Errorf("%s (use -allow-secrets)", why)
		}
	}
	st, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	if !st.Mode().IsRegular() {
		return false, fmt.Errorf("restore needs a regular source file, not a directory or link")
	}
	backup := path + ".bak"
	bst, err := os.Lstat(backup)
	if err != nil {
		return false, err
	}
	if !bst.Mode().IsRegular() {
		return false, fmt.Errorf("backup is not a regular file")
	}
	raw, err := os.ReadFile(backup)
	if err != nil {
		return false, err
	}
	current, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if bytes.Equal(current, raw) {
		return false, nil
	}
	beforeRestore := path + ".restore.bak"
	if _, err := os.Lstat(beforeRestore); err == nil {
		return false, fmt.Errorf("%s already exists; move it aside before restoring again", beforeRestore)
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if !write {
		return true, nil
	}
	if err := textio.WriteExclusive(beforeRestore, current, st.Mode().Perm()); err != nil {
		return false, fmt.Errorf("pre-restore backup failed; source unchanged: %w", err)
	}
	if err := textio.WriteAtomicMode(path, raw, st.Mode().Perm()); err != nil {
		return false, err
	}
	return true, nil
}
