package glossary

import (
	"strings"
	"testing"
)

func TestReflectedDownloadRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"reflected file download attack", "reflected file download risk review", "reflected_file_download_risk_review", 1},
		{"Reflected file download attack", "Reflected file download risk review", "Reflected_file_download_risk_review", 1},
		{"rEfLeCtEd FiLe DoWnLoAd AtTaCk", "reflected file download risk review", "reflected_file_download_risk_review", 1},
		{"REFLECTED FILE DOWNLOAD ATTACK", "REFLECTED FILE DOWNLOAD RISK REVIEW", "REFLECTED_FILE_DOWNLOAD_RISK_REVIEW", 1},
		{"reflected file download attack; reflected file download attack", "reflected file download risk review; reflected file download risk review", "reflected_file_download_risk_review; reflected_file_download_risk_review", 2},
		{"review:\nreflected file download attack", "review:\nreflected file download risk review", "review:\nreflected_file_download_risk_review", 1},
		{"(reflected file download attack).", "(reflected file download risk review).", "(reflected_file_download_risk_review).", 1},
		{"跨站脚本包含攻击", "跨站脚本包含数据泄露风险核查", "跨站脚本包含数据泄露风险核查", 1},
		{"cross site script inclusion attack", "cross site script inclusion data exposure risk review", "cross_site_script_inclusion_data_exposure_risk_review", 1},
		{"open redirect attack", "untrusted destination redirect risk review", "untrusted_destination_redirect_risk_review", 1},
	}
	for _, in := range []string{
		"reflected file download RFD download file attachment filename",
		"Content-Disposition Content-Type filename*=UTF-8 X-Content-Type-Options nosniff",
		"reflected file download prevention validateFilename setContentDisposition",
		"reflected file download risk review",
		"反射型文件下载攻击 反射文件下载攻击",
		"RFD attack file download attack",
		"reflected-file-download-attack reflected_file_download_attack",
		"reflected file download attacks",
		"reflected\tfile download attack",
		"reflected  file download attack",
		"reflected\nfile download attack",
		"reflected file download\tattack",
		"xreflected file download attack",
		"reflected file download attackx",
		"2reflected file download attack",
		"reflected file download attack2",
		"_reflected file download attack",
		"reflected file download attack_suffix",
		"前reflected file download attack",
		"reflected file download attack后",
		`import "reflected file download attack"`,
		"https://example.invalid/reflected%20file%20download%20attack",
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
