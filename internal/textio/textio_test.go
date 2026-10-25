package textio

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDecodeEncode(t *testing.T) {
	cases := []struct {
		name     string
		raw      string
		wantText string
		wantFmt  Format
		// wantBack is what Encode(Decode(raw)) yields; "" means raw itself.
		wantBack string
		// wantStrayCRLF allows a CRLF in the decoded text. Decode drops only
		// the CR of a CRLF pair, so a lone CR in front of one stays content
		// and leaves "\r\n" behind. Every other input must decode CR-free.
		wantStrayCRLF bool
	}{
		{name: "empty", raw: "", wantText: ""},
		{name: "plain LF", raw: "a\nb\n", wantText: "a\nb\n"},
		{name: "no final newline", raw: "a\nb", wantText: "a\nb"},
		{
			name: "BOM only", raw: "\xEF\xBB\xBF", wantText: "",
			wantFmt: Format{BOM: true},
		},
		{
			name: "BOM and LF", raw: "\xEF\xBB\xBFa\nb\n", wantText: "a\nb\n",
			wantFmt: Format{BOM: true},
		},
		{
			name: "CRLF", raw: "a\r\nb\r\n", wantText: "a\nb\n",
			wantFmt: Format{CRLF: true},
		},
		{
			name: "CRLF without final newline", raw: "a\r\nb", wantText: "a\nb",
			wantFmt: Format{CRLF: true},
		},
		{
			name: "BOM and CRLF", raw: "\xEF\xBB\xBFx\r\n\r\ny\r\n", wantText: "x\n\ny\n",
			wantFmt: Format{BOM: true, CRLF: true},
		},
		{
			// Every CRLF becomes LF; the majority style is restored.
			name: "mixed mostly CRLF", raw: "a\r\nb\nc\r\n", wantText: "a\nb\nc\n",
			wantFmt: Format{CRLF: true}, wantBack: "a\r\nb\r\nc\r\n",
		},
		{
			name: "mixed mostly LF", raw: "a\nb\r\nc\n", wantText: "a\nb\nc\n",
			wantFmt: Format{}, wantBack: "a\nb\nc\n",
		},
		{
			// The CRLF is rewritten even though LF wins the vote.
			name: "mixed one CRLF of four", raw: "a\r\nb\nc\nd\n", wantText: "a\nb\nc\nd\n",
			wantFmt: Format{}, wantBack: "a\nb\nc\nd\n",
		},
		{
			// A tie counts as CRLF.
			name: "mixed tie", raw: "a\r\nb\n", wantText: "a\nb\n",
			wantFmt: Format{CRLF: true}, wantBack: "a\r\nb\r\n",
		},
		{
			name: "lone CR untouched", raw: "a\rb\n", wantText: "a\rb\n",
		},
		{
			// A lone CR in front of a CRLF: only the pair's own CR is
			// dropped, so the decoded text keeps a literal CR plus a newline.
			// Encode rewrites every "\n" unconditionally, so this still
			// round-trips byte for byte. See TestStrayCRRoundTrip.
			name: "lone CR before CRLF", raw: "a\r\r\n", wantText: "a\r\n",
			wantFmt: Format{CRLF: true}, wantStrayCRLF: true,
		},
		{
			name: "BOM not at start is content", raw: "a\xEF\xBB\xBF\n",
			wantText: "a\xEF\xBB\xBF\n",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			text, f := Decode([]byte(tc.raw))
			if string(text) != tc.wantText {
				t.Errorf("Decode text = %q, want %q", text, tc.wantText)
			}
			if f != tc.wantFmt {
				t.Errorf("Decode format = %+v, want %+v", f, tc.wantFmt)
			}
			if got := bytes.Contains(text, []byte("\r\n")); got != tc.wantStrayCRLF {
				t.Errorf("CRLF in decoded text = %v, want %v (text %q)",
					got, tc.wantStrayCRLF, text)
			}
			want := tc.wantBack
			if want == "" {
				want = tc.raw
			}
			if back := Encode(text, f); string(back) != want {
				t.Errorf("Encode = %q, want %q", back, want)
			}
		})
	}
}

func TestEncodeChangedText(t *testing.T) {
	// The scanner hands back shorter text; Encode must still apply the format.
	got := Encode([]byte("x\n\ny"), Format{BOM: true, CRLF: true})
	if want := "\xEF\xBB\xBFx\r\n\r\ny"; string(got) != want {
		t.Errorf("Encode = %q, want %q", got, want)
	}
	if got := Encode([]byte("x\n"), Format{}); string(got) != "x\n" {
		t.Errorf("Encode with zero Format = %q, want unchanged", got)
	}
}

func TestIsBinary(t *testing.T) {
	big := bytes.Repeat([]byte("a"), binarySniffLen)
	cases := []struct {
		name string
		raw  []byte
		want bool
	}{
		{"empty", nil, false},
		{"ascii", []byte("int main() {}\n"), false},
		{"utf8", []byte("// \xE4\xB8\xAD\xE6\x96\x87\n"), false},
		{"gbk", []byte("// \xD6\xD0\xCE\xC4\n"), false},
		{"nul at start", []byte("\x00abc"), true},
		{"nul in middle", []byte("ab\x00c"), true},
		{"utf16le", []byte("a\x00b\x00"), true},
		{"nul just inside the sniff window", append(big[:len(big)-1:len(big)-1], 0), true},
		{"nul after the sniff window", append(append([]byte{}, big...), 0), false},
	}
	for _, tc := range cases {
		if got := IsBinary(tc.raw); got != tc.want {
			t.Errorf("%s: IsBinary = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestWriteAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.go")
	if err := os.WriteFile(path, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomic(path, []byte("new")); err != nil {
		t.Fatalf("WriteAtomic: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("content = %q, want %q", got, "new")
	}
	if runtime.GOOS != "windows" { // Windows only models the read-only bit
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != 0o600 {
			t.Errorf("mode = %v, want 0600 preserved", st.Mode().Perm())
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("temporary files left behind: %v", names)
	}
}

func TestWriteAtomicNewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.go")
	if err := WriteAtomic(path, []byte("x")); err != nil {
		t.Fatalf("WriteAtomic: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "x" {
		t.Errorf("content = %q, want %q", got, "x")
	}
}

func TestWriteAtomicMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.sh")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := WriteAtomicMode(path, []byte("new"), 0o755); err != nil {
		t.Fatalf("WriteAtomicMode: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "new" {
		t.Errorf("content = %q, want %q", got, "new")
	}
	if runtime.GOOS != "windows" { // Windows only models the read-only bit
		st, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if st.Mode().Perm() != 0o755 {
			t.Errorf("mode = %v, want 0755", st.Mode().Perm())
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("directory holds %d entries, want 1", len(entries))
	}
}

// TestStrayCRRoundTrip pins down the one case where a CR survives Decode: a
// lone CR in front of a CRLF is content, so the pair must come back intact.
func TestStrayCRRoundTrip(t *testing.T) {
	for _, raw := range []string{
		"a\r\r\n",
		"a\r\r\nb\r\n",
		"x\ry\r\n",
		"\r\r\n",
		"a\rb",
		"a\r\n\r\r\nb\r\n",
	} {
		text, f := Decode([]byte(raw))
		if got := string(Encode(text, f)); got != raw {
			t.Errorf("round trip of %q = %q", raw, got)
		}
	}
}
