package maildir

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/meysam81/parse-dmarc/internal/config"
	"github.com/rs/zerolog"
)

const (
	reportName = "google.com!example.com!1609459200!1609545600.xml"
	reportXML  = `<?xml version="1.0"?><feedback><report_metadata><report_id>1</report_id></report_metadata></feedback>`
)

// reportMessage is a message as the MTA delivers it: a text body plus the
// aggregate report attachment.
func reportMessage() string {
	return "From: noreply-dmarc-support@google.com\r\n" +
		"To: dmarc@example.com\r\n" +
		"Subject: Report Domain: example.com Submitter: google.com\r\n" +
		"Message-ID: <report-1@google.com>\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=\"b\"\r\n\r\n" +
		"--b\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n" +
		"Report attached.\r\n" +
		"--b\r\nContent-Type: application/xml; name=\"" + reportName + "\"\r\n" +
		"Content-Disposition: attachment; filename=\"" + reportName + "\"\r\n\r\n" +
		reportXML + "\r\n--b--\r\n"
}

// newMaildir returns a Maildir skeleton with the given files in new/.
func newMaildir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for _, sub := range []string{"cur", "new", "tmp"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, "new", name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestFetchThenMoveToCur(t *testing.T) {
	dir := newMaildir(t, map[string]string{
		"1700000000.report":  reportMessage(),
		"1700000001.chatter": "From: someone@example.com\r\nSubject: no report here\r\n\r\nhello\r\n",
	})

	log := zerolog.Nop()
	c := NewClient(&config.MaildirConfig{Path: dir}, &log)

	result, err := c.FetchDMARCReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Reports) != 1 || len(result.MessageIDs) != 2 {
		t.Fatalf("want 1 report out of 2 messages, got %d reports / %d message IDs", len(result.Reports), len(result.MessageIDs))
	}
	if len(result.Reports[0].Attachments) != 1 {
		t.Fatalf("want 1 attachment, got %d", len(result.Reports[0].Attachments))
	}
	att := result.Reports[0].Attachments[0]
	if att.Filename != reportName || string(att.Data) != reportXML {
		t.Fatalf("want the report attachment %q, got %q (%d bytes)", reportName, att.Filename, len(att.Data))
	}
	if result.Reports[0].MessageID != "<report-1@google.com>" {
		t.Fatalf("want the Message-ID header, got %q", result.Reports[0].MessageID)
	}

	if err := c.MoveToCur(result.MessageIDs); err != nil {
		t.Fatal(err)
	}

	newLeft, err := os.ReadDir(filepath.Join(dir, "new"))
	if err != nil {
		t.Fatal(err)
	}
	if len(newLeft) != 0 {
		t.Fatalf("want new/ empty after the move, %d left", len(newLeft))
	}
	// Both messages move, report or not: on disk new/ is the queue.
	for _, name := range []string{"1700000000.report:2,S", "1700000001.chatter:2,S"} {
		if _, err := os.Stat(filepath.Join(dir, "cur", name)); err != nil {
			t.Errorf("want %s in cur/: %v", name, err)
		}
	}
	content, err := os.ReadFile(filepath.Join(dir, "cur", "1700000000.report:2,S"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "<feedback>") {
		t.Error("message body changed on the way to cur/")
	}

	// A moved batch is not fetched twice.
	again, err := c.FetchDMARCReports()
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Reports) != 0 || len(again.MessageIDs) != 0 {
		t.Fatalf("want an empty second fetch, got %d reports / %d message IDs", len(again.Reports), len(again.MessageIDs))
	}
}

func TestFetchWithoutNewDir(t *testing.T) {
	log := zerolog.Nop()
	c := NewClient(&config.MaildirConfig{Path: t.TempDir()}, &log)
	if _, err := c.FetchDMARCReports(); err == nil {
		t.Fatal("want an error for a directory that is not a Maildir")
	}
}

func TestCurNameReplacesExistingInfo(t *testing.T) {
	if got := curName("1700000000.report:2,SR"); got != "1700000000.report:2,S" {
		t.Fatalf("want the info block replaced, got %q", got)
	}
	if got := curName("1700000000.report"); got != "1700000000.report:2,S" {
		t.Fatalf("want an info block added, got %q", got)
	}
}
