package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAdditionalLanguagesCLI(t *testing.T) {
	files := map[string]string{
		"player.gd":          "var x = \"# text\" # gone\n",
		"material.gdshader":  "shader_type spatial; // gone\n",
		"common.gdshaderinc": "float f() { return 1.0; } /* gone */\n",
		"compute.wgsl":       "let x = 1; /* outer /* nested */ tail */\n",
		"meson.build":        "project('a#b') # gone\n",
		"meson.options":      "option('feature') # gone\n",
		"meson_options.txt":  "option('legacy') # gone\n",
		"BUILD":              "x = \"# text\" # gone\n",
		"WORKSPACE":          "workspace(name = 'demo') # gone\n",
		"MODULE.bazel":       "module(name = 'demo') # gone\n",
		"rules.bzl":          "x = 1 # gone\n",
	}
	want := map[string]string{
		"player.gd":          "var x = \"# text\"\n",
		"material.gdshader":  "shader_type spatial;\n",
		"common.gdshaderinc": "float f() { return 1.0; }\n",
		"compute.wgsl":       "let x = 1;\n",
		"meson.build":        "project('a#b')\n",
		"meson.options":      "option('feature')\n",
		"meson_options.txt":  "option('legacy')\n",
		"BUILD":              "x = \"# text\"\n",
		"WORKSPACE":          "workspace(name = 'demo')\n",
		"MODULE.bazel":       "module(name = 'demo')\n",
		"rules.bzl":          "x = 1\n",
	}
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
}
