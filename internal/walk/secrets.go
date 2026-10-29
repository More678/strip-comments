package walk

import (
	"path/filepath"
	"strings"
)

// Files that normally hold credentials are skipped by default: a tool that
// rewrites files in place has no business editing a private key. Pass
// -allow-secrets to process them anyway.
//
// A file is only skipped when it also resolves to a supported language, so
// this changes nothing for the many credential files the tool already ignores
// for lack of a language (a bare `.env`, `id_rsa`). It matters for the ones
// that do look like source: `credentials.json`, `secrets.yml`, `.npmrc`.

// secretNames are matched against the lower-cased base name.
var secretNames = []string{
	".env", ".envrc", ".netrc", "_netrc", ".npmrc", ".pypirc", ".dockercfg",
	"credentials", "credentials.json", "credentials.yml", "credentials.yaml",
	"secrets.json", "secrets.yml", "secrets.yaml", "secret.json",
	"id_rsa", "id_dsa", "id_ecdsa", "id_ed25519", "identity",
	"kubeconfig", ".kubeconfig", "known_hosts", "authorized_keys",
	".htpasswd", ".pgpass", ".my.cnf", ".git-credentials",
	"service-account.json", "serviceaccount.json", "gha-creds.json",
	"terraform.tfvars", "terraform.tfstate", ".terraformrc",
}

// secretGlobs are matched against the lower-cased base name.
var secretGlobs = []string{
	"*.pem", "*.key", "*.pfx", "*.p12", "*.jks", "*.keystore", "*.truststore",
	"*.asc", "*.gpg", "*.kdbx", "*.ppk", "*.crt", "*.cer", "*.der",
	".env.*", "*.env", "env.*.local", "*_rsa", "*_dsa", "*_ed25519",
	"*secret*.json", "*secret*.yml", "*secret*.yaml",
	"*credential*.json", "*credential*.yml", "*credential*.yaml",
	"*.tfvars", "*.tfvars.json",
}

// secretDirs are path segments that put everything below them off limits.
var secretDirs = []string{
	".ssh", ".aws", ".gnupg", ".kube", ".docker", ".azure", ".gcloud",
	"secrets", ".secrets",
}

// SecretReason names why path looks like a credential file, or "" when it does
// not. The reason is shown to the user, so it says which rule matched.
func SecretReason(path string) string {
	base := strings.ToLower(filepath.Base(path))
	for _, n := range secretNames {
		if base == n {
			return "credential file (" + n + ")"
		}
	}
	for _, g := range secretGlobs {
		if ok, err := filepath.Match(g, base); err == nil && ok {
			return "credential file (" + g + ")"
		}
	}
	slashed := strings.ToLower(filepath.ToSlash(path))
	for _, d := range secretDirs {
		if strings.Contains(slashed, "/"+d+"/") || strings.HasPrefix(slashed, d+"/") {
			return "inside " + d + "/"
		}
	}
	return ""
}
