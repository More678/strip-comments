package scan

import (
	"strings"
	"testing"

	"github.com/More678/strip-comments/internal/lang"
)

func TestModernRustCSharpLiterals(t *testing.T) {
	cases := []struct {
		name, language, in, want string
		unsafe                   bool
	}{
		{"rust raw c", "rust", `let s = cr#"a" /* keep */ "b"#; // gone` + "\n", `let s = cr#"a" /* keep */ "b"#;` + "\n", false},
		{"rust raw c hashes", "rust", `let s = cr###"a"## // keep` + "\n" + `/* keep */"###; // gone` + "\n", `let s = cr###"a"## // keep` + "\n" + `/* keep */"###;` + "\n", false},
		{"rust raw c no hash", "rust", `let s = cr"a\"; // gone` + "\n", `let s = cr"a\";` + "\n", false},
		{"rust raw byte", "rust", `let s = br#"a" /* keep */ "b"#; // gone` + "\n", `let s = br#"a" /* keep */ "b"#;` + "\n", false},
		{"rust raw c unclosed", "rust", "// gone\nlet s = cr##\"a\"#; /* keep */\n", "// gone\nlet s = cr##\"a\"#; /* keep */\n", true},
		{"csharp raw triple", "csharp", `var s = """a " /* keep */ b"""; // gone` + "\n", `var s = """a " /* keep */ b""";` + "\n", false},
		{"csharp raw four", "csharp", `var s = """"a """ /* keep */ """ b""""; // gone` + "\n", `var s = """"a """ /* keep */ """ b"""";` + "\n", false},
		{"csharp raw five", "csharp", "var s = \"\"\"\"\"\na \"\"\"\" // keep\n/* keep */\n\"\"\"\"\"; // gone\n", "var s = \"\"\"\"\"\na \"\"\"\" // keep\n/* keep */\n\"\"\"\"\";\n", false},
		{"csharp raw unclosed", "csharp", "// gone\nvar s = \"\"\"\"a \"\"\" /* keep */\n", "// gone\nvar s = \"\"\"\"a \"\"\" /* keep */\n", true},
		{"csharp nested string", "csharp", `var s = $"{ "a /* keep */ b" }"; // gone` + "\n", `var s = $"{ "a /* keep */ b" }";` + "\n", false},
		{"csharp expression comment", "csharp", `var s = $"{1 /* gone */}"; // gone` + "\n", `var s = $"{1 }";` + "\n", false},
		{"csharp nested interpolation", "csharp", `var s = $"{ $"{1 /* gone */} /* keep */" }"; // gone` + "\n", `var s = $"{ $"{1 } /* keep */" }";` + "\n", false},
		{"csharp verbatim interpolation", "csharp", `var s = $@"a "" /* keep */ { "// keep" /* gone */ }"; // gone` + "\n", `var s = $@"a "" /* keep */ { "// keep" }";` + "\n", false},
		{"csharp alternate verbatim", "csharp", `var s = @$"a {1 /* gone */} /* keep */"; // gone` + "\n", `var s = @$"a {1 } /* keep */";` + "\n", false},
		{"csharp escaped braces", "csharp", `var s = $"{{ /* keep */ }} {1 /* gone */}"; // gone` + "\n", `var s = $"{{ /* keep */ }} {1 }";` + "\n", false},
		{"csharp raw interpolation", "csharp", `var s = $$"""{ /* keep */ } {{ "a /* keep */" /* gone */ }}"""; // gone` + "\n", `var s = $$"""{ /* keep */ } {{ "a /* keep */" }}""";` + "\n", false},
		{"csharp initializer", "csharp", `var s = $"{new int[] { 1 /* gone */ }[0]}"; // gone` + "\n", `var s = $"{new int[] { 1 }[0]}";` + "\n", false},
		{"csharp format text", "csharp", `var s = $"{1:/* keep */}"; // gone` + "\n", `var s = $"{1:/* keep */}";` + "\n", false},
		{"csharp interpolation unclosed", "csharp", "// keep\nvar s = $\"{1 /* keep */\n", "// keep\nvar s = $\"{1 /* keep */\n", true},
		{"csharp expression line comment", "csharp", "var s = $\"{1 // gone\n}\"; // gone\n", "var s = $\"{1\n}\";\n", false},
		{"csharp char brace", "csharp", `var s = $"{'}' /* gone */}"; // gone` + "\n", `var s = $"{'}' }";` + "\n", false},
		{"csharp conditional", "csharp", `var s = $"{(true ? "/* keep */" : "// keep") /* gone */}";` + "\n", `var s = $"{(true ? "/* keep */" : "// keep") }";` + "\n", false},
		{"csharp alias", "csharp", `var s = $"{global::System.Int32.MaxValue /* gone */}";` + "\n", `var s = $"{global::System.Int32.MaxValue }";` + "\n", false},
		{"csharp verbatim format", "csharp", `var s = $@"{1:/* keep */}"; // gone` + "\n", `var s = $@"{1:/* keep */}";` + "\n", false},
		{"csharp raw format", "csharp", `var s = $$"""{{1:/* keep */}}"""; // gone` + "\n", `var s = $$"""{{1:/* keep */}}""";` + "\n", false},
		{"csharp unclosed comment", "csharp", `var s = $"{1 /* keep }";`, `var s = $"{1 /* keep }";`, true},
		{"csharp unbalanced expression", "csharp", `var s = $"{(1]}";`, `var s = $"{(1]}";`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := lang.ByName(tc.language)
			res := Strip([]byte(tc.in), l, Options{})
			if string(res.Output) != tc.want || res.Unsafe != tc.unsafe {
				t.Fatalf("got %q unsafe=%v (%s); want %q unsafe=%v", res.Output, res.Unsafe, res.Reason, tc.want, tc.unsafe)
			}
			if !tc.unsafe {
				again := Strip(res.Output, l, Options{})
				if again.Unsafe || string(again.Output) != tc.want || again.Comments != 0 {
					t.Fatalf("second pass changed output: %+v", again)
				}
			}
		})
	}
}

func TestRustCSharpKeepDoc(t *testing.T) {
	for _, language := range []string{"rust", "csharp"} {
		t.Run(language, func(t *testing.T) {
			const src = "/// documentation\n/* remove */\nint_value; // remove\n"
			for _, keep := range []bool{false, true} {
				want := "int_value;\n"
				if keep {
					want = "/// documentation\n" + want
				}
				res := Strip([]byte(src), lang.ByName(language), Options{KeepDoc: keep})
				if res.Unsafe || string(res.Output) != want {
					t.Fatalf("keep-doc=%v: got %+v, want %q", keep, res, want)
				}
			}
		})
	}
}

func TestCSharpStringNestingLimit(t *testing.T) {
	src := "// keep\n" + strings.Repeat(`$"{`, 130) + "1" + strings.Repeat(`}"`, 130) + ";\n"
	res := Strip([]byte(src), lang.ByName("csharp"), Options{})
	if !res.Unsafe || string(res.Output) != src || !strings.Contains(res.Reason, "nesting limit") {
		t.Fatalf("expected unchanged input at nesting limit, got %+v", res)
	}
}
