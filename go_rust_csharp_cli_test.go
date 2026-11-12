package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoRustCSharpCLI(t *testing.T) {
	files := map[string]string{
		"sample.go": "package demo\n// documentation\nvar text = `https://example.invalid/* literal */` // remove\nvar n = 1 /* remove */ + 2\n",
		"sample.rs": "/// documentation\nfn main() {\n    let text = r##\"https://example.invalid/* literal */\"##; // remove\n    let n = 1 /* outer /* nested */ tail */ + 2;\n}\n",
		"sample.cs": "/// documentation\nclass Demo {\n    string text = @\"https://example.invalid/\"\"/* literal */\"; // remove\n    int/**/number = 1; /* remove */\n}\n",
	}
	want := map[string]string{
		"sample.go": "package demo\nvar text = `https://example.invalid/* literal */`\nvar n = 1 + 2\n",
		"sample.rs": "fn main() {\n    let text = r##\"https://example.invalid/* literal */\"##;\n    let n = 1 + 2;\n}\n",
		"sample.cs": "class Demo {\n    string text = @\"https://example.invalid/\"\"/* literal */\";\n    int number = 1;\n}\n",
	}
	files["modern.rs"] = "fn main() { let s = cr#\"a\" /* literal */ \"b\"#; } // remove\n"
	want["modern.rs"] = "fn main() { let s = cr#\"a\" /* literal */ \"b\"#; }\n"
	files["modern.cs"] = "class Modern { string s = \"\"\"\"a \"\"\" /* literal */ b\"\"\"\"; string t = $\"{1 /* remove */}\"; } // remove\n"
	want["modern.cs"] = "class Modern { string s = \"\"\"\"a \"\"\" /* literal */ b\"\"\"\"; string t = $\"{1 }\"; }\n"
	dir := srcDir(t, files)
	for _, write := range []bool{false, true} {
		args := []string{dir}
		expected := files
		if write {
			args = []string{"-w", "-backup", dir}
			expected = want
		}
		if code, out := runCLI(t, args...); code != exitOK {
			t.Fatalf("write=%v exit=%d: %s", write, code, out)
		}
		for name, text := range expected {
			got, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil || string(got) != text {
				t.Errorf("write=%v %s: %q, %v; want %q", write, name, got, err, text)
			}
			if write {
				backup, err := os.ReadFile(filepath.Join(dir, name+".bak"))
				if err != nil || string(backup) != files[name] {
					t.Errorf("%s backup differs: %q, %v", name, backup, err)
				}
			}
		}
	}
	if code, out := runCLI(t, "-check", dir); code != exitOK {
		t.Fatalf("second pass exit=%d: %s", code, out)
	}
}

func TestLanguageBoundariesCLI(t *testing.T) {
	cases := []struct {
		name, file, in, want string
	}{
		{"go directive", "sample.go", "package demo\n//go:generate not-run\n// remove\nvar n = 1\n", "package demo\n//go:generate not-run\nvar n = 1\n"},
		{"rust doc boundary", "sample.rs", "/// docs\n/** block docs */\n/*** ordinary */\n//// ordinary\nfn main() {}\n", "/// docs\n/** block docs */\nfn main() {}\n"},
		{"csharp format boundary", "sample.cs", `var s = $"{1:F"; var other = "}"; // keep` + "\n", `var s = $"{1:F"; var other = "}"; // keep` + "\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := srcDir(t, map[string]string{tc.file: tc.in})
			if code, out := runCLI(t, "-w", "-backup", "-keep-doc", dir); code != exitOK {
				t.Fatalf("exit=%d: %s", code, out)
			}
			path := filepath.Join(dir, tc.file)
			got, err := os.ReadFile(path)
			if err != nil || string(got) != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
			backup, err := os.ReadFile(path + ".bak")
			if tc.in == tc.want {
				if !os.IsNotExist(err) {
					t.Fatalf("skipped input created a backup: %q, %v", backup, err)
				}
			} else if err != nil || string(backup) != tc.in {
				t.Fatalf("backup differs: %q, %v", backup, err)
			}
			if code, out := runCLI(t, "-check", "-keep-doc", dir); code != exitOK {
				t.Fatalf("second pass exit=%d: %s", code, out)
			}
		})
	}
}
