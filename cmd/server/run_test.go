package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dmarcguardhq/parse-dmarc/internal/config"
	"github.com/dmarcguardhq/parse-dmarc/internal/logger"
	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
)

const goodReportXML = `<?xml version="1.0"?><feedback><report_metadata><org_name>google.com</org_name>` +
	`<report_id>1</report_id><date_range><begin>1609459200</begin><end>1609545600</end></date_range>` +
	`</report_metadata><policy_published><domain>example.com</domain><p>none</p></policy_published>` +
	`<record><row><source_ip>192.0.2.1</source_ip><count>1</count><policy_evaluated><disposition>none</disposition>` +
	`</policy_evaluated></row><identifiers><header_from>example.com</header_from></identifiers></record></feedback>`

// maildirMessage is a message as the MTA delivers it: a text body and, when
// name is not empty, one attachment.
func maildirMessage(name, content string) string {
	msg := "From: noreply-dmarc-support@google.com\r\n" +
		"Subject: Report Domain: example.com\r\n" +
		"Message-ID: <" + name + "@google.com>\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=\"b\"\r\n\r\n" +
		"--b\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\nReport attached.\r\n"
	if name != "" {
		msg += "--b\r\nContent-Type: application/xml; name=\"" + name + "\"\r\n" +
			"Content-Disposition: attachment; filename=\"" + name + "\"\r\n\r\n" + content + "\r\n"
	}
	return msg + "--b--\r\n"
}

func newTestMaildir(t *testing.T, files map[string]string) string {
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

// A message no report can be read from — no attachment, or an attachment that
// does not parse — ends up in the deadletter mailbox, while a good report is
// done with in cur/ and in the store.
func TestFetchMaildirReportsDeadlettersUnreadableMessage(t *testing.T) {
	log = logger.NewLogger("error", false)

	dir := newTestMaildir(t, map[string]string{
		"1.report":   maildirMessage("google.com!example.com!1609459200.xml", goodReportXML),
		"2.broken":   maildirMessage("broken.xml", "<not-a-dmarc-report/>"),
		"3.chatter":  maildirMessage("", ""),
		"4.also-bad": maildirMessage("truncated.gz", "PK not a report"),
	})

	store, err := storage.NewStorage(filepath.Join(t.TempDir(), "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()

	cfg := &config.Config{Maildir: config.MaildirConfig{Path: dir, DeadletterMailbox: "DEADLETTER"}}
	if err := fetchMaildirReports(cfg, store, nil); err != nil {
		t.Fatal(err)
	}

	newLeft, err := os.ReadDir(filepath.Join(dir, "new"))
	if err != nil {
		t.Fatal(err)
	}
	if len(newLeft) != 0 {
		t.Fatalf("want new/ empty after the fetch, %d left", len(newLeft))
	}

	if _, err := os.Stat(filepath.Join(dir, "cur", "1.report:2,S")); err != nil {
		t.Errorf("want the parsed report in cur/: %v", err)
	}
	for _, name := range []string{"2.broken:2,S", "3.chatter:2,S", "4.also-bad:2,S"} {
		if _, err := os.Stat(filepath.Join(dir, "DEADLETTER", name)); err != nil {
			t.Errorf("want %s in the deadletter mailbox, not dropped: %v", name, err)
		}
	}
}
