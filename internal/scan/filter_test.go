package scan

import (
	"testing"

	"github.com/More678/Remove-All-Commet/internal/lang"
)

// TestKeepLicenseRun covers the header run that KeepLicense protects: it
// starts at the first comment of the file and stops at the first blank line.
func TestKeepLicenseRun(t *testing.T) {
	cases := []struct {
		name     string
		lang     string
		in       string
		want     string
		comments int
	}{
		{
			name:     "three-line // header kept whole",
			lang:     "c",
			in:       "// Copyright 2024 A\n// Licensed under X\n// See LICENSE\nint a; // gone\n",
			want:     "// Copyright 2024 A\n// Licensed under X\n// See LICENSE\nint a;\n",
			comments: 1,
		},
		{
			name:     "blank line ends the header",
			lang:     "c",
			in:       "// Copyright 2024 A\n// Licensed under X\n\n// not the licence\nint a;\n",
			want:     "// Copyright 2024 A\n// Licensed under X\n\nint a;\n",
			comments: 1,
		},
		{
			name:     "single block header behaves as before",
			lang:     "c",
			in:       "/* Copyright 2024 */\nint a; // gone\n",
			want:     "/* Copyright 2024 */\nint a;\n",
			comments: 1,
		},
		{
			name:     "three-line # header kept whole",
			lang:     "python",
			in:       "# Copyright 2024 A\n# Licensed under X\n# See LICENSE\nx = 1  # gone\n",
			want:     "# Copyright 2024 A\n# Licensed under X\n# See LICENSE\nx = 1\n",
			comments: 1,
		},
		{
			name:     "header run after a shebang",
			lang:     "python",
			in:       "#!/usr/bin/env python\n# Copyright 2024 A\n# Licensed under X\n\n# gone\nx = 1\n",
			want:     "#!/usr/bin/env python\n# Copyright 2024 A\n# Licensed under X\n\nx = 1\n",
			comments: 1,
		},
		{
			name:     "multi-line block header kept whole",
			lang:     "c",
			in:       "/*\n * Copyright 2024 A\n * Licensed under X\n */\nint a; // gone\n",
			want:     "/*\n * Copyright 2024 A\n * Licensed under X\n */\nint a;\n",
			comments: 1,
		},
		{
			name:     "no header when code comes first",
			lang:     "c",
			in:       "int a;\n// Copyright 2024 A\n// Licensed under X\n",
			want:     "int a;\n",
			comments: 2,
		},
		{
			name:     "run does not reach past a blank line to a doc comment",
			lang:     "c",
			in:       "// Copyright 2024 A\n\n/** doc */\nint a;\n",
			want:     "// Copyright 2024 A\n\nint a;\n",
			comments: 1,
		},
	}

	for _, tc := range cases {
		l := lang.ByName(tc.lang)
		if l == nil {
			t.Errorf("%s: language %q is not registered", tc.name, tc.lang)
			continue
		}
		got := Strip([]byte(tc.in), l, Options{KeepLicense: true})
		if string(got.Output) != tc.want {
			t.Errorf("%s: output mismatch\n in:   %q\n got:  %q\n want: %q",
				tc.name, tc.in, got.Output, tc.want)
		}
		if got.Comments != tc.comments {
			t.Errorf("%s: Comments = %d, want %d", tc.name, got.Comments, tc.comments)
		}
	}
}

// TestKeepLicenseOffByDefault makes sure the header is only protected when the
// option asks for it.
func TestKeepLicenseOffByDefault(t *testing.T) {
	in := []byte("// Copyright 2024 A\n// Licensed under X\nint a;\n")
	got := Strip(in, lang.ByName("c"), Options{})
	if string(got.Output) != "int a;\n" {
		t.Errorf("got %q, want %q", got.Output, "int a;\n")
	}
}
