package glossary

import (
	"strings"
	"testing"
)

func TestUnrestrictedUploadRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"不受限文件上传漏洞", "不受限文件上传风险核查", "不受限文件上传风险核查", 1},
		{"unrestricted file upload vulnerability", "unrestricted file upload risk review", "unrestricted_file_upload_risk_review", 1},
		{"Unrestricted file upload vulnerability", "Unrestricted file upload risk review", "Unrestricted_file_upload_risk_review", 1},
		{"uNrEsTrIcTeD FiLe UpLoAd VuLnErAbIlItY", "unrestricted file upload risk review", "unrestricted_file_upload_risk_review", 1},
		{"UNRESTRICTED FILE UPLOAD VULNERABILITY", "UNRESTRICTED FILE UPLOAD RISK REVIEW", "UNRESTRICTED_FILE_UPLOAD_RISK_REVIEW", 1},
		{"不受限文件上传漏洞; unrestricted file upload vulnerability", "不受限文件上传风险核查; unrestricted file upload risk review", "不受限文件上传风险核查; unrestricted_file_upload_risk_review", 2},
		{"review:\nunrestricted file upload vulnerability", "review:\nunrestricted file upload risk review", "review:\nunrestricted_file_upload_risk_review", 1},
		{"(unrestricted file upload vulnerability).", "(unrestricted file upload risk review).", "(unrestricted_file_upload_risk_review).", 1},
		{"reflected file download attack", "reflected file download risk review", "reflected_file_download_risk_review", 1},
		{"Zip Slip attack", "archive extraction path traversal risk review", "archive_extraction_path_traversal_risk_review", 1},
	}
	for _, in := range []string{
		"unrestricted file upload 不受限文件上传 upload file multipart/form-data",
		"Content-Type filename FormData File CWE-434",
		"unrestricted file upload prevention validateFileType scanUpload allowlist",
		"unrestricted file upload risk review 不受限文件上传风险核查",
		"不受限文件上传防护 不受限 文件上传漏洞 不受限文件上传 漏洞",
		"arbitrary file upload vulnerability 任意文件上传漏洞",
		"file upload vulnerability unrestricted upload vulnerability",
		"unrestricted-file-upload-vulnerability unrestricted_file_upload_vulnerability",
		"unrestricted file upload vulnerabilities",
		"unrestricted\tfile upload vulnerability",
		"unrestricted  file upload vulnerability",
		"unrestricted\nfile upload vulnerability",
		"unrestricted file upload\tvulnerability",
		"xunrestricted file upload vulnerability",
		"unrestricted file upload vulnerabilityx",
		"2unrestricted file upload vulnerability",
		"unrestricted file upload vulnerability2",
		"_unrestricted file upload vulnerability",
		"unrestricted file upload vulnerability_suffix",
		"前unrestricted file upload vulnerability",
		"unrestricted file upload vulnerability后",
		`import "unrestricted file upload vulnerability"`,
		`import "不受限文件上传漏洞"`,
		"https://example.invalid/unrestricted%20file%20upload%20vulnerability",
		"https://example.invalid/不受限文件上传漏洞",
	} {
		cases = append(cases, struct {
			in, prose, code string
			hits            int
		}{in, in, in, 0})
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			for _, comment := range []bool{false, true} {
				opt := Options{}
				want := tc.code
				if comment {
					opt.Comments = []Region{{Start: 0, End: len(tc.in), Comment: true}}
					want = tc.prose
				}
				out, hits := tb.Apply([]byte(tc.in), opt)
				if string(out) != want || len(hits) != tc.hits {
					t.Fatalf("comment=%v: got %q (%d hits), want %q (%d hits)", comment, out, len(hits), want, tc.hits)
				}
				for _, h := range hits {
					if h.Start < 0 || h.End > len(tc.in) || h.Start >= h.End {
						t.Fatalf("invalid hit range: %+v", h)
					}
					if h.InComment != comment || h.Line != strings.Count(tc.in[:h.Start], "\n")+1 || tc.in[h.Start:h.End] != h.From || !strings.Contains(want, h.To) {
						t.Errorf("incorrect hit metadata: %+v", h)
					}
				}
				if comment {
					opt.Comments[0].End = len(out)
				}
				again, secondHits := tb.Apply(out, opt)
				if string(again) != string(out) || len(secondHits) != 0 {
					t.Fatalf("second pass changed output: %q (%+v)", again, secondHits)
				}
			}
		})
	}
}
