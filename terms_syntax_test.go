package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestReplacedCodeStillCompiles is the test that matters most for -replace:
// a rewritten file has to stay syntactically valid. A multi-word replacement
// landing on an identifier is the way that breaks, so each fixture puts a term
// in a function name, a variable and a comment at once.
//
// Each case is skipped when its checker is not installed, so the suite still
// runs on a machine without a C compiler.
func TestReplacedCodeStillCompiles(t *testing.T) {
	cases := []struct {
		name, file, body string
		tool             string
		args             []string // the file name is appended
	}{
		{
			name: "c", file: "a.c", tool: "gcc", args: []string{"-fsyntax-only"},
			body: "/* 渗透测试工具 */\n#include <stdio.h>\n" +
				"int exploit(void) {        // 漏洞利用入口\n" +
				"    int payload = 1;       /* 攻击载荷 */\n" +
				"    printf(\"%d\\n\", payload);\n" +
				"    return 0;              // 提权\n}\n",
		},
		{
			name: "python", file: "b.py", tool: "python", args: []string{"-m", "py_compile"},
			body: "#!/usr/bin/env python3\n# 渗透测试脚本\n" +
				"def exploit():             # 漏洞利用\n" +
				"    payload = 1            # 攻击载荷\n" +
				"    return payload\n",
		},
		{
			name: "shell", file: "c.sh", tool: "sh", args: []string{"-n"},
			body: "#!/bin/sh\n# 渗透测试\nexploit() {   # 漏洞利用\n" +
				"    payload=1 # 攻击载荷\n    echo \"$payload\"\n}\nexploit\n",
		},
		{
			name: "go", file: "d.go", tool: "gofmt", args: []string{"-e"},
			body: "package x\n\n// 渗透测试入口\nfunc Exploit() int {  // 漏洞利用\n" +
				"\tpayload := 1      /* 攻击载荷 */\n\treturn payload\n}\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bin, err := exec.LookPath(tc.tool)
			if err != nil {
				t.Skipf("%s is not installed", tc.tool)
			}
			dir := srcDir(t, map[string]string{tc.file: tc.body})
			path := filepath.Join(dir, tc.file)

			code, out := runCLI(t, "-replace", "-w", "-q", dir)
			if code != exitOK {
				t.Fatalf("exit = %d:\n%s", code, out)
			}
			after, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(after) == tc.body {
				t.Fatalf("nothing was replaced in %s", tc.file)
			}

			cmd := exec.Command(bin, append(tc.args, path)...)
			cmd.Dir = dir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("%s rejected the rewritten file: %v\n%s\n--- file:\n%s",
					tc.tool, err, out, after)
			}
		})
	}
}
