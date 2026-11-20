package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testHost = "127.0.0.1:38123"

func record(h http.Handler, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func previewPOST(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "http://"+testHost+"/api/preview", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestEmbeddedRoutesEP(t *testing.T) {
	h := newHandler(testHost)
	for path, kind := range map[string]string{"/": "text/html", "/app.css": "text/css", "/app.js": "javascript", "/api/languages": "application/json"} {
		w := record(h, httptest.NewRequest(http.MethodGet, "http://"+testHost+path, nil))
		if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Type"), kind) || w.Body.Len() == 0 {
			t.Fatalf("GET %s: status %d, content type %q, %d bytes", path, w.Code, w.Header().Get("Content-Type"), w.Body.Len())
		}
		if path != "/api/languages" {
			continue
		}
		var langs []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &langs); err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for _, entry := range langs {
			name, named := entry["name"].(string)
			risky, flagged := entry["risky"].(bool)
			if _, duplicate := seen[name]; duplicate || !named || name == "" || !flagged {
				t.Fatalf("invalid or duplicate language entry: %v", entry)
			}
			seen[name] = risky
		}
		if risky, found := seen["go"]; !found || risky || !seen["matlab"] {
			t.Errorf("expected go (regular) and matlab (heuristic), got %v", seen)
		}
	}
}

func TestLocalOriginDecisionTable(t *testing.T) {
	h := newHandler(testHost)
	for _, tc := range []struct {
		name, host, origin string
		status             int
	}{
		{"no origin", testHost, "", 200},
		{"same origin", testHost, "http://" + testHost, 200},
		{"foreign host", "example.invalid:38123", "", 403},
		{"other local port", "127.0.0.1:38124", "", 403},
		{"foreign origin", testHost, "https://example.invalid", 403},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := previewPOST(`{"language":"go","source":""}`)
			r.Host = tc.host
			if tc.origin != "" {
				r.Header.Set("Origin", tc.origin)
			}
			if w := record(h, r); w.Code != tc.status {
				t.Fatalf("status %d, want %d", w.Code, tc.status)
			}
		})
	}
}

func TestRejectedPreviewEP(t *testing.T) {
	const media = "application/json"
	const valid = `{"language":"go","source":"PRIVATE_PREVIEW_MARKER"}`
	h := newHandler(testHost)
	for _, tc := range []struct {
		name, media, body string
		status            int
	}{
		{"missing content type", "", valid, 415},
		{"other content type", "text/plain", valid, 415},
		{"non-UTF-8 charset", "application/json; charset=iso-8859-1", valid, 415},
		{"unknown field", media, `{"language":"go","source":"","extra":true}`, 400},
		{"trailing value", media, valid + ` {}`, 400},
		{"missing source", media, `{"language":"go"}`, 400},
		{"wrong source type", media, `{"language":"go","source":5}`, 400},
		{"invalid UTF-8", media, `{"language":"go","source":"` + "\xff" + `"}`, 400},
		{"NUL in text", media, `{"language":"go","source":"PRIVATE_PREVIEW_MARKER\u0000"}`, 400},
		{"protected basename", media, `{"language":"go","source":"","filename":".env"}`, 403},
		{"path instead of basename", media, `{"language":"go","source":"","filename":"src/demo.go"}`, 400},
		{"Windows path", media, `{"language":"go","source":"","filename":"src\\demo.go"}`, 400},
		{"unknown language", media, `{"language":"not-a-language","source":""}`, 400},
		{"auto with no match", media, `{"language":"auto","source":"","filename":"demo.unknown"}`, 400},
		{"auto empty paste", media, `{"language":"auto","source":""}`, 400},
		{"auto ambiguous snippet", media, `{"language":"auto","source":"x = 1\n"}`, 400},
		{"auto plain text", media, `{"language":"auto","source":"Release notes for this week.\n"}`, 400},
		{"auto heuristic requires explicit choice", media, `{"language":"auto","source":"","filename":"demo.asm"}`, 400},
		{"keep without replacement", media, `{"language":"go","source":"","keepComments":true}`, 400},
		{"unsafe unterminated comment", media, `{"language":"go","source":"package demo\n/* PRIVATE_PREVIEW_MARKER"}`, 422},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := previewPOST(tc.body)
			r.Header.Del("Content-Type")
			if tc.media != "" {
				r.Header.Set("Content-Type", tc.media)
			}
			w := record(h, r)
			var obj map[string]any
			if w.Code != tc.status || json.Unmarshal(w.Body.Bytes(), &obj) != nil {
				t.Fatalf("status %d, want %d with a JSON error: %.160s", w.Code, tc.status, w.Body.String())
			}
			msg, _ := obj["error"].(string)
			_, output := obj["output"]
			_, diff := obj["diff"]
			if msg == "" || output || diff || strings.Contains(w.Body.String(), "PRIVATE_PREVIEW_MARKER") {
				t.Fatal("response must report an error without returning source, output, or diff")
			}
		})
	}
}

func TestLanguageDetectionDecisionTable(t *testing.T) {
	const goSource = `"source":"package main\nfunc main() {}\n"`
	h := newHandler(testHost)
	for _, tc := range []struct{ name, body, lang, by string }{
		{"pasted Go", `{"language":"auto",` + goSource + `}`, "go", "content"},
		{"unrecognized suffix falls back to content", `{"language":"auto","filename":"demo.unknown",` + goSource + `}`, "go", "content"},
		{"extensionless Python script with BOM", `{"language":"auto","filename":"run","source":"` + string(rune(0xfeff)) + `#!/usr/bin/env python3\nprint(1)\n"}`, "python", "shebang"},
		{"filename takes priority over content", `{"language":"auto","filename":"demo.py",` + goSource + `}`, "python", "filename"},
		{"manual choice takes priority over filename", `{"language":"go","filename":"demo.py","source":"x = 1\n"}`, "go", "manual"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := record(h, previewPOST(tc.body))
			var got map[string]any
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &got) != nil {
				t.Fatalf("status %d; expected JSON result: %.160s", w.Code, w.Body.String())
			}
			if got["language"] != tc.lang || got["detectedBy"] != tc.by {
				t.Errorf("got %v by %v, want %s by %s", got["language"], got["detectedBy"], tc.lang, tc.by)
			}
		})
	}
}

func TestPreviewOptionsDecisionTable(t *testing.T) {
	const code = "package demo\n" + `var label = "Find audit logs by HTTP method; 渗透攻击"` + "\n"
	const comment = "// Find audit logs by HTTP method; 渗透攻击\n"
	const newCode = "package demo\n" + `var label = "Find_audit_log_entries_by_HTTP_method; 渗透攻击"` + "\n"
	const newComment = "// Find audit log entries by HTTP method; 渗透攻击\n"
	const mixed = "\xEF\xBB\xBFpackage demo\r\nvar n = 1\n"
	h := newHandler(testHost)
	for _, tc := range []struct {
		name, source, want, lang string
		options                  map[string]any
		comments, removed, terms int
		omitDiff                 bool
	}{
		{"default glossary off", code + comment, code, "go", nil, 1, 1, 0, false},
		{"full audit glossary only", code + comment, newCode + newComment, "go", map[string]any{"replace": true, "keepComments": true}, 0, 0, 2, false},
		{"only kept comments", code + comment, code + newComment, "go", map[string]any{"replace": true, "keepComments": true, "commentsOnly": true}, 0, 0, 1, false},
		{"auto preserves BOM and CRLF", "\xEF\xBB\xBFpackage demo\r\n// note\r\n", "\xEF\xBB\xBFpackage demo\r\n", "go", map[string]any{"language": "auto", "filename": "demo.go"}, 1, 1, 0, false},
		{"unchanged mixed endings", mixed, mixed, "go", nil, 0, 0, 0, false},
		{"empty source", "", "", "go", nil, 0, 0, 0, false},
		{"empty output", "// note\n", "", "go", nil, 1, 1, 0, false},
		{"retain banner docs and blank line", "/* license */\nint a;\n/** docs */\n// note\nint b;\n", "/* license */\nint a;\n/** docs */\n\nint b;\n", "c", map[string]any{"keepLicense": true, "keepDoc": true, "keepBlank": true}, 1, 0, 0, false},
		{"large diff omitted but output retained", strings.Repeat("n = 1 # note\n", 6000), strings.Repeat("n = 1\n", 6000), "python", nil, 6000, 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := map[string]any{"source": tc.source, "language": tc.lang}
			for k, v := range tc.options {
				request[k] = v
			}
			body, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			r := previewPOST(string(body))
			r.Header.Set("Content-Type", "application/json; charset=UTF-8")
			w := record(h, r)
			var got map[string]any
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &got) != nil {
				t.Fatalf("status %d; expected JSON result: %.160s", w.Code, w.Body.String())
			}
			changed := tc.source != tc.want
			for key, want := range map[string]any{"output": tc.want, "language": tc.lang, "changed": changed, "diffOmitted": tc.omitDiff,
				"comments": float64(tc.comments), "removedLines": float64(tc.removed), "terms": float64(tc.terms)} {
				if got[key] != want {
					t.Errorf("%s = %.160v, want %.160v", key, got[key], want)
				}
			}
			diff, ok := got["diff"].(string)
			if !ok || (changed && !tc.omitDiff && !strings.Contains(diff, "@@")) || ((!changed || tc.omitDiff) && diff != "") {
				t.Errorf("unexpected diff: %.160q", diff)
			}
		})
	}
}

func TestSourceByteLimitBVA(t *testing.T) {
	const limit = 256 << 10
	atLimit := strings.Repeat("é", limit/len("é"))
	h := newHandler(testHost)
	for _, tc := range []struct {
		name, suffix string
		status       int
	}{{"at byte limit", "", 200}, {"one byte over", "a", 413}} {
		t.Run(tc.name, func(t *testing.T) {
			w := record(h, previewPOST(`{"language":"c","source":"`+atLimit+tc.suffix+`"}`))
			if w.Code != tc.status {
				t.Errorf("%d source bytes: status %d, want %d", len(atLimit)+len(tc.suffix), w.Code, tc.status)
			}
		})
	}
}
