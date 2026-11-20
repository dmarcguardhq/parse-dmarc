package logger

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l, err := newLogger(&buf, "info", FormatJSON, true)
	if err != nil {
		t.Fatal(err)
	}
	l.Info().Str("domain", "example.com").Msg("saved report")
	l.Debug().Msg("dropped below the level")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d: %q", len(lines), buf.String())
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("not JSON: %v: %s", err, lines[0])
	}
	if entry["level"] != "info" || entry["message"] != "saved report" || entry["domain"] != "example.com" || entry["time"] == nil {
		t.Errorf("entry = %v", entry)
	}
}

func TestUnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	l, err := newLogger(&buf, "info", "xml", true)
	if err == nil {
		t.Fatal("want an error for an unknown format")
	}
	l.Info().Msg("still logs")
	if !strings.Contains(buf.String(), "still logs") {
		t.Errorf("console fallback wrote %q", buf.String())
	}
}
