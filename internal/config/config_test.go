package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/caarlos0/env/v11"
)

// A Maildir is the intake instead of a mailbox, so IMAP credentials are not
// required when one is configured.
func TestValidateMaildirNeedsNoIMAP(t *testing.T) {
	cfg := &Config{Maildir: MaildirConfig{Path: "/srv/mail/dmarc/Maildir"}}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("a configured Maildir must validate without IMAP settings, got %v", err)
	}

	if err := (&Config{}).Validate(); err == nil {
		t.Fatal("want an error when neither a Maildir nor an IMAP host is configured")
	}
}

func TestMaildirPathFromEnv(t *testing.T) {
	t.Setenv("MAILDIR_PATH", "/srv/mail/dmarc/Maildir")

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Maildir.Path != "/srv/mail/dmarc/Maildir" {
		t.Fatalf("want MAILDIR_PATH in the config, got %q", cfg.Maildir.Path)
	}
}

func TestMaildirDeadletterBoxFromEnv(t *testing.T) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Maildir.DeadletterMailbox != "DEADLETTER" {
		t.Fatalf("want the IMAP_DEADLETTER_BOX default, got %q", cfg.Maildir.DeadletterMailbox)
	}

	t.Setenv("IMAP_DEADLETTER_BOX", "BROKEN")
	if err := env.Parse(&cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.Maildir.DeadletterMailbox != "BROKEN" {
		t.Fatalf("want IMAP_DEADLETTER_BOX in the config, got %q", cfg.Maildir.DeadletterMailbox)
	}
}

// The dashboard serves the domains and sending hosts collected from the
// mailbox, so it stays on the machine that runs it unless an operator says
// otherwise. The container images set SERVER_HOST because the port mapping
// reaches the interface, not loopback.
func TestServerHostDefaultsToLoopback(t *testing.T) {
	t.Setenv("DATABASE_PATH", filepath.Join(t.TempDir(), "db.sqlite"))
	t.Setenv("SERVER_HOST", "")

	cfg, err := Load(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Host != DefaultServerHost {
		t.Fatalf("want the %s default, got %q", DefaultServerHost, cfg.Server.Host)
	}
}

func TestServerHostFromEnv(t *testing.T) {
	t.Setenv("DATABASE_PATH", filepath.Join(t.TempDir(), "db.sqlite"))
	t.Setenv("SERVER_HOST", "0.0.0.0")

	cfg, err := Load(filepath.Join(t.TempDir(), "absent.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("want SERVER_HOST to win over the default, got %q", cfg.Server.Host)
	}
}

// GenerateSample writes a file an operator copies, and the file overrides the
// environment. Pinning the address there would beat SERVER_HOST in silence, so
// the sample leaves the key out and the default still comes from Load.
func TestGenerateSampleLeavesHostToTheEnvironment(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := GenerateSample(path); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var sample struct {
		Server map[string]any `json:"server"`
	}
	if err := json.Unmarshal(data, &sample); err != nil {
		t.Fatal(err)
	}
	if _, ok := sample.Server["host"]; ok {
		t.Fatalf("a sample that names the host would override SERVER_HOST, got:\n%s", data)
	}

	t.Setenv("DATABASE_PATH", filepath.Join(dir, "db.sqlite"))
	t.Setenv("SERVER_HOST", "0.0.0.0")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Fatalf("want SERVER_HOST to reach the generated sample, got %q", cfg.Server.Host)
	}
}
