package glossary

import (
	"strings"
	"testing"
)

func TestTLSHostnameRiskTerms(t *testing.T) {
	tb := NewTable(Default)
	cases := []struct {
		in, prose, code string
		hits            int
	}{
		{"TLS证书主机名验证漏洞", "TLS证书主机名验证风险核查", "TLS证书主机名验证风险核查", 1},
		{"TLS certificate hostname validation vulnerability", "TLS certificate hostname validation risk review", "TLS_certificate_hostname_validation_risk_review", 1},
		{"tls certificate hostname validation vulnerability", "TLS certificate hostname validation risk review", "TLS_certificate_hostname_validation_risk_review", 1},
		{"Tls certificate hostname validation vulnerability", "TLS certificate hostname validation risk review", "TLS_certificate_hostname_validation_risk_review", 1},
		{"tLs CeRtIfIcAtE HoStNaMe VaLiDaTiOn VuLnErAbIlItY", "TLS certificate hostname validation risk review", "TLS_certificate_hostname_validation_risk_review", 1},
		{"TLS CERTIFICATE HOSTNAME VALIDATION VULNERABILITY", "TLS CERTIFICATE HOSTNAME VALIDATION RISK REVIEW", "TLS_CERTIFICATE_HOSTNAME_VALIDATION_RISK_REVIEW", 1},
		{"TLS证书主机名验证漏洞; TLS certificate hostname validation vulnerability", "TLS证书主机名验证风险核查; TLS certificate hostname validation risk review", "TLS证书主机名验证风险核查; TLS_certificate_hostname_validation_risk_review", 2},
		{"review:\nTLS certificate hostname validation vulnerability", "review:\nTLS certificate hostname validation risk review", "review:\nTLS_certificate_hostname_validation_risk_review", 1},
		{"(TLS certificate hostname validation vulnerability).", "(TLS certificate hostname validation risk review).", "(TLS_certificate_hostname_validation_risk_review).", 1},
		{"TLS协议降级攻击", "TLS协议降级风险核查", "TLS协议降级风险核查", 1},
		{"TLS protocol downgrade attack", "TLS protocol downgrade risk review", "TLS_protocol_downgrade_risk_review", 1},
		{"IDN homograph attack", "IDN homograph domain spoofing risk review", "IDN_homograph_domain_spoofing_risk_review", 1},
	}
	for _, in := range []string{
		"TLS X.509 SAN DNS-ID hostname validation TLS certificate hostname validation TLS证书主机名验证",
		"VerifyHostname ServerName HostnameVerifier InsecureSkipVerify CWE-297",
		"TLS certificate hostname validation prevention validateCertificate trustStore",
		"TLS certificate hostname validation risk review TLS证书主机名验证风险核查",
		"TLS证书主机名验证防护 tls证书主机名验证漏洞 Tls证书主机名验证漏洞 TLS 证书主机名验证漏洞 TLS证书主机名验证 漏洞",
		"certificate hostname validation vulnerability 证书主机名验证漏洞",
		"TLS certificate hostname verification vulnerability TLS hostname validation vulnerability",
		"TLS-certificate-hostname-validation-vulnerability TLS_certificate_hostname_validation_vulnerability",
		"TLS certificate hostname validation vulnerabilities",
		"TLS\tcertificate hostname validation vulnerability",
		"TLS  certificate hostname validation vulnerability",
		"TLS\ncertificate hostname validation vulnerability",
		"TLS certificate hostname validation\tvulnerability",
		"xTLS certificate hostname validation vulnerability",
		"TLS certificate hostname validation vulnerabilityx",
		"2TLS certificate hostname validation vulnerability",
		"TLS certificate hostname validation vulnerability2",
		"_TLS certificate hostname validation vulnerability",
		"TLS certificate hostname validation vulnerability_suffix",
		"前TLS certificate hostname validation vulnerability",
		"TLS certificate hostname validation vulnerability后",
		`import "TLS certificate hostname validation vulnerability"`,
		`import "TLS证书主机名验证漏洞"`,
		"https://example.invalid/TLS%20certificate%20hostname%20validation%20vulnerability",
		"https://example.invalid/TLS证书主机名验证漏洞",
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
