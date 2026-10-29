package walk

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretReason(t *testing.T) {
	secret := []string{
		".env", ".env.production", "prod.env", ".envrc",
		"credentials.json", "secrets.yml", "secrets.yaml", "secret.json",
		"my-secret.yaml", "app-credentials.yml",
		"id_rsa", "deploy_rsa", "id_ed25519", "server.pem", "app.key",
		"store.jks", "cert.p12", "key.ppk", "backup.kdbx",
		".npmrc", ".pypirc", ".netrc", ".pgpass", ".htpasswd",
		".git-credentials", "kubeconfig", "authorized_keys",
		"terraform.tfvars", "prod.tfvars", "service-account.json",
	}
	for _, name := range secret {
		if SecretReason(name) == "" {
			t.Errorf("SecretReason(%q) = \"\", want a reason", name)
		}
	}

	// Ordinary source must stay processable, or the guard would make the
	// tool useless on a normal project.
	ordinary := []string{
		"main.go", "config.yml", "app.yaml", "package.json", "tsconfig.json",
		"settings.py", "Makefile", "README.md", "values.yaml",
		"docker-compose.yml", "keyboard.go", "monkey.py", "certificate.md",
	}
	for _, name := range ordinary {
		if why := SecretReason(name); why != "" {
			t.Errorf("SecretReason(%q) = %q, want \"\"", name, why)
		}
	}
}

func TestSecretReasonDirs(t *testing.T) {
	for _, p := range []string{
		filepath.Join(".ssh", "config.yml"),
		filepath.Join("home", "u", ".aws", "conf.json"),
		filepath.Join("srv", ".kube", "x.yaml"),
		filepath.Join("repo", "secrets", "db.yml"),
	} {
		if SecretReason(p) == "" {
			t.Errorf("SecretReason(%q) = \"\", want a reason", p)
		}
	}
	// A directory that merely contains the word is not a match.
	for _, p := range []string{
		filepath.Join("src", "secretsmanager", "client.go"),
		filepath.Join("app", "sshd", "main.go"),
	} {
		if why := SecretReason(p); why != "" {
			t.Errorf("SecretReason(%q) = %q, want \"\"", p, why)
		}
	}
}

// TestWalkSkipsSecrets checks the wiring: a credential file that does resolve
// to a language is reported as a skip, and -allow-secrets returns it.
func TestWalkSkipsSecrets(t *testing.T) {
	root := t.TempDir()
	for _, f := range []string{"secrets.yml", "app.yml", "prod.tfvars"} {
		write(t, root, f, "x: 1 # c\n")
	}

	files, skips, errs := Walk([]string{root}, Options{})
	if len(errs) != 0 {
		t.Fatalf("errors: %v", errs)
	}
	if got := relPaths(t, root, files); !equalPaths(got, []string{"app.yml"}) {
		t.Errorf("files = %v, want [app.yml]", got)
	}
	if len(skips) != 2 {
		t.Fatalf("skips = %+v, want 2", skips)
	}
	for _, s := range skips {
		if !strings.Contains(s.Reason, "-allow-secrets") {
			t.Errorf("skip %q does not say how to override", s.Reason)
		}
	}

	files, skips, _ = Walk([]string{root}, Options{AllowSecrets: true})
	if len(files) != 3 || len(skips) != 0 {
		t.Errorf("with AllowSecrets: %d files, %d skips, want 3 and 0",
			len(files), len(skips))
	}
}
