// Package maildir reads DMARC report mail from a local Maildir, for hosts that
// already receive the reports on the filesystem they run the parser on.
package maildir

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emersion/go-message/mail"
	"github.com/meysam81/parse-dmarc/internal/config"
	"github.com/meysam81/parse-dmarc/internal/imap"
	"github.com/rs/zerolog"
)

// Client reads the messages a MTA has delivered into a Maildir.
type Client struct {
	config *config.MaildirConfig
	log    *zerolog.Logger
	// batch holds the file names of the last fetch, in MessageID order, so
	// the move to cur/ names the same files that were read even if the
	// directory gained messages in between.
	batch []string
}

// NewClient creates a new Maildir client. It fails fast when the process
// cannot read the configured path: an unreadable Maildir would otherwise look
// like an empty one on every cycle.
func NewClient(cfg *config.MaildirConfig, log *zerolog.Logger) (*Client, error) {
	if _, err := os.ReadDir(cfg.Path); err != nil {
		return nil, fmt.Errorf("maildir %s: %w", cfg.Path, err)
	}
	return &Client{config: cfg, log: log}, nil
}

// FetchDMARCReports reads every message in new/ and returns the ones carrying a
// DMARC report attachment. MessageIDs are the positions of the messages in this
// batch: on disk the file name is the durable key, not the number.
func (c *Client) FetchDMARCReports() (*imap.FetchResult, error) {
	newDir := filepath.Join(c.config.Path, "new")
	entries, err := os.ReadDir(newDir)
	if err != nil {
		return nil, fmt.Errorf("read maildir %s: %w", newDir, err)
	}

	var result imap.FetchResult
	c.batch = c.batch[:0]
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		path := filepath.Join(newDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			// A message mid-delivery (tmp/ -> new/) or removed by a human
			// happens in a Maildir; skip it and look again next cycle.
			c.log.Warn().Err(err).Str("file", name).Msg("skipping maildir message")
			continue
		}

		id := uint32(len(c.batch) + 1)
		c.batch = append(c.batch, name)

		report := imap.Report{UID: id}
		report.Subject, report.From, report.Date, report.MessageID = c.headers(data)
		report.Attachments = imap.CollectAttachments(c.log, bytes.NewReader(data))

		// A message without a recognized report is a deadletter, not a report:
		// it leaves new/ for the deadletter mailbox, where the operator sees
		// that a sender's report went unrecognized.
		if len(report.Attachments) == 0 {
			c.log.Warn().
				Str("file", name).
				Str("message_id", report.MessageID).
				Str("from", report.From).
				Str("subject", report.Subject).
				Msg("no DMARC report attachment in message; it is moved to the deadletter mailbox")
			result.DeadletterIDs = append(result.DeadletterIDs, id)
			continue
		}

		result.MessageIDs = append(result.MessageIDs, id)
		result.Reports = append(result.Reports, report)
	}

	return &result, nil
}

// MoveToCur moves the messages of the last fetch to cur/, which is how a
// Maildir marks a message as read.
func (c *Client) MoveToCur(ids []uint32) error {
	return c.move(filepath.Join(c.config.Path, "cur"), ids)
}

// MoveToDeadletter moves the messages of the last fetch that no report could
// be read from to the deadletter mailbox, so the operator sees what the
// parser dropped instead of it passing as processed.
func (c *Client) MoveToDeadletter(ids []uint32) error {
	if len(ids) == 0 {
		return nil
	}

	dir := filepath.Join(c.config.Path, c.config.DeadletterMailbox)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create deadletter mailbox %s: %w", dir, err)
	}
	return c.move(dir, ids)
}

// move renames the given messages of the last fetch out of new/ into dir.
func (c *Client) move(dir string, ids []uint32) error {
	var firstErr error
	for _, id := range ids {
		if id < 1 || int(id) > len(c.batch) {
			continue
		}

		name := c.batch[id-1]
		src := filepath.Join(c.config.Path, "new", name)
		dst := filepath.Join(dir, infoName(name))

		if err := os.Rename(src, dst); err != nil {
			if os.IsNotExist(err) {
				c.log.Debug().Str("file", name).Msg("maildir message already moved")
				continue
			}
			c.log.Error().Err(err).Str("file", name).Str("dir", dir).Msg("failed to move maildir message")
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// infoName is the new/ file name with the `:2,S` info suffix a message carries
// once it has been moved out of new/. An existing info block is replaced,
// never stacked.
func infoName(name string) string {
	if i := strings.Index(name, ":2,"); i >= 0 {
		name = name[:i]
	}
	return name + ":2,S"
}

// headers reads the fields the run log quotes, and the Message-ID that finds
// the message again in the MTA log.
func (c *Client) headers(data []byte) (subject, from, date, messageID string) {
	mr, err := mail.CreateReader(bytes.NewReader(data))
	if err != nil {
		c.log.Warn().Err(err).Msg("failed to create mail reader")
		return
	}
	defer func() { _ = mr.Close() }()

	subject = mr.Header.Get("Subject")
	messageID = mr.Header.Get("Message-ID")
	if addrs, err := mr.Header.AddressList("From"); err == nil && len(addrs) > 0 {
		from = addrs[0].Address
	}
	if d, err := mr.Header.Date(); err == nil {
		date = d.String()
	}
	return subject, from, date, messageID
}
