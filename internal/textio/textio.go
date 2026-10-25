// Package textio converts files between their on-disk shape and the plain
// "\n" text the scanner expects, and writes them back atomically.
package textio

import (
	"bytes"
	"os"
	"path/filepath"
)

// bom is the UTF-8 byte order mark.
var bom = []byte{0xEF, 0xBB, 0xBF}

// binarySniffLen is how much of a file is inspected for NUL bytes.
const binarySniffLen = 8 << 10

// Format records the on-disk details that Decode normalises away.
type Format struct {
	BOM  bool // the file started with a UTF-8 BOM
	CRLF bool // most line endings were CRLF
}

// Decode strips a UTF-8 BOM and turns every CRLF into a plain LF, so the
// scanner only ever sees "\n" line endings. Format.CRLF records whether CRLF
// was the majority ending, which is what Encode restores; a mixed file
// therefore comes back in its dominant style.
//
// A lone CR is content, not a line ending, and is left untouched. Only the CR
// of a CRLF pair is dropped, so "a\r\r\n" decodes to "a\r\n": the surviving CR
// is the literal one, and Encode puts the pair back as "a\r\r\n".
func Decode(raw []byte) (text []byte, f Format) {
	text = raw
	if bytes.HasPrefix(text, bom) {
		f.BOM = true
		text = text[len(bom):]
	}
	crlf := bytes.Count(text, []byte("\r\n"))
	if crlf == 0 {
		return text, f
	}
	// A tie counts as CRLF: that keeps the diff small on Windows files.
	f.CRLF = crlf*2 >= bytes.Count(text, []byte("\n"))
	out := make([]byte, 0, len(text)-crlf)
	for {
		i := bytes.IndexByte(text, '\n')
		if i < 0 {
			out = append(out, text...)
			break
		}
		if i > 0 && text[i-1] == '\r' {
			out = append(out, text[:i-1]...)
		} else {
			out = append(out, text[:i]...)
		}
		out = append(out, '\n')
		text = text[i+1:]
	}
	return out, f
}

// Encode restores the BOM and the CRLF line endings recorded in f.
//
// Every "\n" becomes "\r\n" unconditionally. A CR already in front of a "\n"
// is literal content that Decode deliberately kept, so skipping the insertion
// there would swallow that byte and lose it on the round trip.
func Encode(text []byte, f Format) []byte {
	out := text
	if f.CRLF {
		n := bytes.Count(text, []byte("\n"))
		out = make([]byte, 0, len(text)+n)
		for {
			i := bytes.IndexByte(text, '\n')
			if i < 0 {
				out = append(out, text...)
				break
			}
			out = append(out, text[:i]...)
			out = append(out, '\r', '\n')
			text = text[i+1:]
		}
	}
	if f.BOM {
		res := make([]byte, 0, len(bom)+len(out))
		res = append(res, bom...)
		return append(res, out...)
	}
	return out
}

// IsBinary reports whether raw contains a NUL byte in its first 8 KiB.
func IsBinary(raw []byte) bool {
	head := raw
	if len(head) > binarySniffLen {
		head = head[:binarySniffLen]
	}
	return bytes.IndexByte(head, 0) >= 0
}

// WriteAtomic replaces path with data: it writes a temporary file next to the
// original, keeps the original file mode and renames it into place, so a
// crash can never leave a half-written source file behind.
func WriteAtomic(path string, data []byte) error {
	mode := os.FileMode(0o644)
	if st, err := os.Stat(path); err == nil {
		mode = st.Mode().Perm()
	}
	return WriteAtomicMode(path, data, mode)
}

// WriteAtomicMode is WriteAtomic with an explicit file mode.
func WriteAtomicMode(path string, data []byte, mode os.FileMode) error {
	dir, base := filepath.Split(path)
	if dir == "" {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, "."+base+".strip-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(name) // no-op once the rename succeeded
	}()

	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(name, mode); err != nil {
		return err
	}
	return os.Rename(name, path)
}
