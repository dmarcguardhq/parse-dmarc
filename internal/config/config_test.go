package config

import (
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
