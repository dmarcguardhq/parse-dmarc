package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/goccy/go-json"
)

var (
	// ErrMissingIMAPHost is returned when IMAP host is not configured
	ErrMissingIMAPHost = errors.New("IMAP_HOST is required: set via environment variable or config file")
	// ErrMissingIMAPUsername is returned when IMAP username is not configured
	ErrMissingIMAPUsername = errors.New("IMAP_USERNAME is required: set via environment variable or config file")
	// ErrMissingIMAPPassword is returned when IMAP password is not configured
	ErrMissingIMAPPassword = errors.New("IMAP_PASSWORD is required: set via environment variable or config file")
)

// Config holds the application configuration
type Config struct {
	LogLevel    string         `json:"log_level" env:"LOG_LEVEL" envDefault:"info"`
	ColoredLogs bool           `json:"colored_logs" env:"COLORED_LOGS" envDefault:"false"`
	IMAP        IMAPConfig     `json:"imap"`
	Maildir     MaildirConfig  `json:"maildir"`
	Database    DatabaseConfig `json:"database"`
	Server      ServerConfig   `json:"server"`
}

// IMAPConfig holds IMAP server configuration
type IMAPConfig struct {
	Host     string `json:"host" env:"IMAP_HOST"`
	Port     int    `json:"port" env:"IMAP_PORT" envDefault:"993"`
	Username string `json:"username" env:"IMAP_USERNAME"`
	Password string `json:"password" env:"IMAP_PASSWORD"`
	Mailbox  string `json:"mailbox" env:"IMAP_MAILBOX" envDefault:"INBOX"`
	UseTLS   bool   `json:"use_tls" env:"IMAP_USE_TLS" envDefault:"true"`

	// StartTLS dials plaintext (typically port 143) and upgrades the
	// connection with the STARTTLS command. It takes precedence over UseTLS.
	StartTLS bool `json:"starttls" env:"IMAP_STARTTLS"`
	// TLSCAFile is a PEM bundle used to verify the IMAP server certificate,
	// for servers whose certificate is issued by an internal CA.
	TLSCAFile string `json:"tls_ca_file" env:"IMAP_TLS_CA_FILE"`
	// TLSSkipVerify disables IMAP server certificate verification. Prefer
	// TLSCAFile; this leaves the connection open to interception.
	TLSSkipVerify bool `json:"tls_skip_verify" env:"IMAP_TLS_SKIP_VERIFY"`

	MarkAsSeen       bool   `json:"mark_as_seen" env:"IMAP_MARK_AS_SEEN" envDefault:"true"`
	ProcessedMailbox string `json:"processed_mailbox" env:"IMAP_PROCESSED_MAILBOX"`
}

// MaildirConfig holds local Maildir intake configuration. When Path is set,
// reports are read from <Path>/new and moved to <Path>/cur instead of being
// fetched over IMAP, and no IMAP credentials are needed.
type MaildirConfig struct {
	Path string `json:"path" env:"MAILDIR_PATH"`
	// DeadletterMailbox is the subdirectory of Path that receives every
	// message no report could be read from, so the operator sees what the
	// parser dropped instead of it passing as processed.
	DeadletterMailbox string `json:"deadletter_mailbox" env:"IMAP_DEADLETTER_BOX" envDefault:"DEADLETTER"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Path string `json:"path" env:"DATABASE_PATH"`
}

// DefaultServerHost is the interface the dashboard binds when SERVER_HOST is
// left unset. Loopback keeps the dashboard on the machine that runs it, because
// the reports it serves name the domains and sending hosts collected from the
// mailbox. Reaching it from elsewhere is then a deliberate act:
// SERVER_HOST=0.0.0.0, which the container images set for the port mapping to
// work.
const DefaultServerHost = "127.0.0.1"

// ServerConfig holds web server configuration
type ServerConfig struct {
	Port int `json:"port" env:"SERVER_PORT" envDefault:"8080"`
	// Host is omitted from a sample config on purpose: the JSON file overrides
	// the environment, so pinning it there would silently beat SERVER_HOST.
	Host string `json:"host,omitempty" env:"SERVER_HOST" envDefault:""`
}

func defaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", errors.New("cannot determine home directory")
	}
	return filepath.Join(home, ".parse-dmarc/db.sqlite"), nil
}

func fallbackDBPath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.New("cannot determine home directory or current working directory")
	}
	return filepath.Join(cwd, ".parse-dmarc/db.sqlite"), nil
}

func ensureDBPathExists(dbPath string) error {
	parent := filepath.Dir(dbPath)
	if err := os.MkdirAll(parent, 0755); err != nil {
		return errors.New("failed to create database directory at " + parent + ": " + err.Error() + " - ensure the path is writable or set DATABASE_PATH environment variable")
	}
	return nil
}

// Load loads configuration from a JSON file
func Load(path string) (*Config, error) {
	var cfg Config
	var err error

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}

	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read config file %s: %w", path, err)
		}

		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config file %s: %w", path, err)
		}
	}

	if cfg.IMAP.Port == 0 {
		cfg.IMAP.Port = 993
	}
	if cfg.IMAP.Mailbox == "" {
		cfg.IMAP.Mailbox = "INBOX"
	}
	if cfg.Maildir.DeadletterMailbox == "" {
		cfg.Maildir.DeadletterMailbox = "DEADLETTER"
	}
	if cfg.Database.Path == "" {
		cfg.Database.Path, err = defaultDBPath()
		if err != nil || ensureDBPathExists(cfg.Database.Path) != nil {
			cfg.Database.Path, err = fallbackDBPath()
			if err != nil {
				return nil, fmt.Errorf("resolve database path: %w", err)
			}
			err = ensureDBPathExists(cfg.Database.Path)
			if err != nil {
				return nil, fmt.Errorf("ensure database path: %w", err)
			}
		}
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = DefaultServerHost
	}

	return &cfg, nil
}

// Validate checks that all required configuration values are set.
// Required fields: IMAP host, username, and password, unless a Maildir
// replaces the mailbox and the fetch does not touch IMAP at all.
// Returns nil if valid, or an error describing the missing configuration.
func (c *Config) Validate() error {
	if c.Maildir.Path != "" {
		return nil
	}
	if c.IMAP.Host == "" {
		return ErrMissingIMAPHost
	}
	if c.IMAP.Username == "" {
		return ErrMissingIMAPUsername
	}
	if c.IMAP.Password == "" {
		return ErrMissingIMAPPassword
	}
	return nil
}

// GenerateSample creates a sample configuration file
func GenerateSample(path string) error {
	dbPath, err := defaultDBPath()
	if err != nil {
		return fmt.Errorf("resolve default database path: %w", err)
	}
	sample := Config{
		LogLevel: "info",
		Maildir:  MaildirConfig{Path: "", DeadletterMailbox: "DEADLETTER"},
		IMAP: IMAPConfig{
			Host:     "imap.example.com",
			Port:     993,
			Username: "your-email@example.com",
			Password: "your-password",
			Mailbox:  "INBOX",
			UseTLS:   true,

			MarkAsSeen:       true,
			ProcessedMailbox: "",
		},
		Database: DatabaseConfig{
			Path: dbPath,
		},
		Server: ServerConfig{
			Port: 8080,
		},
	}

	data, err := json.MarshalIndent(sample, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal sample config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config file %s: %w", path, err)
	}

	return nil
}
