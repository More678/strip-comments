package scan

import (
	"testing"

	"github.com/More678/strip-comments/internal/lang"
)

func TestAdditionalLanguages(t *testing.T) {
	cases := []struct {
		name, language, in, want string
		opt                      Options
		unsafe                   bool
	}{
		{"gdscript strings", "gdscript", "var s = \"# text\" # gone\nvar t = '# text' # gone\n", "var s = \"# text\"\nvar t = '# text'\n", Options{}, false},
		{"gdscript multiline", "gdscript", "var s = \"\"\"a\n# text\n\"\"\" # gone\n", "var s = \"\"\"a\n# text\n\"\"\"\n", Options{}, false},
		{"gdscript raw", "gdscript", `var s = r"a\"# text" # gone` + "\n", `var s = r"a\"# text"` + "\n", Options{}, false},
		{"gdscript doc", "gdscript", "## Documentation\n# gone\nvar x = 1\n", "## Documentation\nvar x = 1\n", Options{KeepDoc: true}, false},
		{"gdscript doc removed", "gdscript", "## Documentation\nvar x = 1\n", "var x = 1\n", Options{}, false},
		{"gdshader include", "gdshader", "#include \"res://shared.gdshaderinc\" // gone\nfloat/**/x = 1.0; /* gone */\n", "#include \"res://shared.gdshaderinc\"\nfloat x = 1.0;\n", Options{}, false},
		{"gdshader doc", "gdshader", "/** uniform docs */\nuniform float x; // gone\n", "/** uniform docs */\nuniform float x;\n", Options{KeepDoc: true}, false},
		{"gdshader empty block", "gdshader", "/**/float x; /* gone */\n", "float x;\n", Options{}, false},
		{"wgsl nested", "wgsl", "let/* outer /* inner */ outer */x = 1; // gone\n", "let x = 1;\n", Options{}, false},
		{"wgsl quoted delimiters", "wgsl", "/* \" /* inner */ \" */\nlet x = 1;\n", "let x = 1;\n", Options{}, false},
		{"wgsl slash inside block", "wgsl", "/* // /* inner */ outer */\nlet x = 1;\n", "let x = 1;\n", Options{}, false},
		{"wgsl empty", "wgsl", "/**/let x = 1;\n", "let x = 1;\n", Options{}, false},
		{"wgsl outer unclosed", "wgsl", "/* outer /* inner */\nlet x = 1;\n", "/* outer /* inner */\nlet x = 1;\n", Options{}, true},
		{"meson strings", "meson", "x = 'a#b' # gone\ny = f'@name@#' # gone\n", "x = 'a#b'\ny = f'@name@#'\n", Options{}, false},
		{"meson raw triple", "meson", "x = '''a\n# text\n\\''' # gone\n# gone\n", "x = '''a\n# text\n\\'''\n", Options{}, false},
		{"meson unclosed triple", "meson", "x = '''a\n# text\n", "x = '''a\n# text\n", Options{}, true},
		{"starlark strings", "starlark", "x = \"# text\" # gone\ny = '''# text\nmore''' # gone\n", "x = \"# text\"\ny = '''# text\nmore'''\n", Options{}, false},
		{"starlark raw", "starlark", `x = r'a\'# text' # gone` + "\n", `x = r'a\'# text'` + "\n", Options{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := lang.ByName(tc.language)
			if l == nil {
				t.Fatalf("unregistered language %s", tc.language)
			}
			res := Strip([]byte(tc.in), l, tc.opt)
			if string(res.Output) != tc.want || res.Unsafe != tc.unsafe {
				t.Fatalf("got %q, unsafe=%v (%s); want %q, unsafe=%v", res.Output, res.Unsafe, res.Reason, tc.want, tc.unsafe)
			}
			if !tc.unsafe {
				again := Strip(res.Output, l, tc.opt)
				if string(again.Output) != tc.want || again.Comments != 0 {
					t.Fatalf("second pass changed output: %+v", again)
				}
			}
		})
	}
}
