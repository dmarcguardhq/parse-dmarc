package imap

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/meysam81/parse-dmarc/internal/config"
	"github.com/rs/zerolog"
)

// Client represents an IMAP client
type Client struct {
	config *config.IMAPConfig
	client *client.Client
	log    *zerolog.Logger
}

// NewClient creates a new IMAP client
func NewClient(cfg *config.IMAPConfig, log *zerolog.Logger) *Client {
	return &Client{config: cfg, log: log}
}

// tlsConfig builds the TLS settings used for both implicit TLS and STARTTLS.
func (c *Client) tlsConfig() (*tls.Config, error) {
	cfg := &tls.Config{
		ServerName:         c.config.Host,
		InsecureSkipVerify: c.config.TLSSkipVerify, // #nosec G402 -- opt-in, for internal IMAP servers
		MinVersion:         tls.VersionTLS12,
	}

	if c.config.TLSCAFile == "" {
		return cfg, nil
	}

	pem, err := os.ReadFile(c.config.TLSCAFile)
	if err != nil {
		return nil, fmt.Errorf("read IMAP CA file %s: %w", c.config.TLSCAFile, err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("no certificates found in IMAP CA file %s", c.config.TLSCAFile)
	}
	cfg.RootCAs = pool

	return cfg, nil
}

// dial opens the connection: implicit TLS, plaintext upgraded via STARTTLS, or
// plaintext. StartTLS wins over UseTLS since UseTLS defaults to true.
func (c *Client) dial(addr string, tlsCfg *tls.Config) (*client.Client, error) {
	if c.config.UseTLS && !c.config.StartTLS {
		return client.DialTLS(addr, tlsCfg)
	}

	imapClient, err := client.Dial(addr)
	if err != nil || !c.config.StartTLS {
		return imapClient, err
	}

	if err := imapClient.StartTLS(tlsCfg); err != nil {
		_ = imapClient.Logout()
		return nil, fmt.Errorf("starttls: %w", err)
	}

	return imapClient, nil
}

// Connect establishes connection to IMAP server
func (c *Client) Connect() error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	c.log.Debug().Str("addr", addr).Msg("connecting")

	tlsCfg, err := c.tlsConfig()
	if err != nil {
		return err
	}

	imapClient, err := c.dial(addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.client = imapClient
	c.log.Info().Str("addr", addr).Msg("connected")

	// Login
	if err := c.client.Login(c.config.Username, c.config.Password); err != nil {
		_ = c.client.Logout()
		return fmt.Errorf("login failed: %w", err)
	}

	c.log.Info().Str("username", c.config.Username).Msg("logged in")
	return nil
}

// Disconnect closes the IMAP connection
func (c *Client) Disconnect() error {
	if c.client != nil {
		return c.client.Logout()
	}
	return nil
}

// Report represents a DMARC report email
type Report struct {
	Subject     string
	From        string
	Date        string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename string
	Data     []byte
}

// FetchResult holds the fetched reports and the message sequence numbers
type FetchResult struct {
	Reports    []Report
	MessageIDs []uint32
}

// FetchDMARCReports fetches DMARC reports from the mailbox
func (c *Client) FetchDMARCReports() (*FetchResult, error) {
	// Select mailbox
	mbox, err := c.client.Select(c.config.Mailbox, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select mailbox: %w", err)
	}

	if mbox.Messages == 0 {
		c.log.Info().Msg("no messages in mailbox")
		return &FetchResult{}, nil
	}

	// Search for unseen messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}

	ids, err := c.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	if len(ids) == 0 {
		c.log.Info().Msg("no new messages found")
		return &FetchResult{}, nil
	}

	c.log.Info().Int("count", len(ids)).Msg("found new messages")

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(ids...)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	section := &imap.BodySectionName{Peek: !c.config.MarkAsSeen}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchFlags}

	go func() {
		done <- c.client.Fetch(seqSet, items, messages)
	}()

	var reports []Report
	var messageIDs []uint32

	for msg := range messages {
		messageIDs = append(messageIDs, msg.SeqNum)

		r := msg.GetBody(section)
		if r == nil {
			c.log.Warn().Uint32("uid", msg.Uid).Msg("server didn't return message body")
			continue
		}

		report := Report{
			Subject: msg.Envelope.Subject,
			Date:    msg.Envelope.Date.String(),
		}

		if len(msg.Envelope.From) > 0 {
			report.From = msg.Envelope.From[0].Address()
		}

		report.Attachments = c.collectAttachments(r, 0)

		// Only add reports with attachments
		if len(report.Attachments) > 0 {
			reports = append(reports, report)
		}
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	return &FetchResult{Reports: reports, MessageIDs: messageIDs}, nil
}

// MarkAsSeen marks messages as seen
func (c *Client) MarkAsSeen(messageIDs []uint32) error {
	if len(messageIDs) == 0 {
		return nil
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(messageIDs...)

	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}

	return c.client.Store(seqSet, item, flags, nil)
}

// MoveMessages moves messages to a destination mailbox using COPY + DELETE + EXPUNGE
func (c *Client) MoveMessages(messageIDs []uint32, destMailbox string) error {
	if len(messageIDs) == 0 {
		return nil
	}

	// Ensure destination mailbox exists (ignore error if it already exists)
	if err := c.client.Create(destMailbox); err != nil {
		c.log.Debug().Err(err).Str("mailbox", destMailbox).Msg("create mailbox (may already exist)")
	}

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(messageIDs...)

	// Copy to destination mailbox
	if err := c.client.Copy(seqSet, destMailbox); err != nil {
		return fmt.Errorf("copy to %s: %w", destMailbox, err)
	}

	// Mark originals as deleted
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.DeletedFlag}
	if err := c.client.Store(seqSet, item, flags, nil); err != nil {
		return fmt.Errorf("mark deleted: %w", err)
	}

	// Expunge deleted messages
	if err := c.client.Expunge(nil); err != nil {
		return fmt.Errorf("expunge: %w", err)
	}

	return nil
}

// maxNestDepth bounds recursion into forwarded/embedded messages.
const maxNestDepth = 4

// collectAttachments walks a MIME message and returns DMARC report attachments,
// descending into embedded messages (Google reports relayed through Exchange or
// Outlook arrive as a message/rfc822 attachment wrapping the real .zip).
func (c *Client) collectAttachments(r io.Reader, depth int) []Attachment {
	if depth > maxNestDepth {
		c.log.Warn().Msg("max nesting depth reached; skipping embedded message")
		return nil
	}

	mr, err := mail.CreateReader(r)
	if err != nil {
		c.log.Warn().Err(err).Msg("failed to create mail reader")
		return nil
	}
	defer func() { _ = mr.Close() }()

	var attachments []Attachment
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			c.log.Warn().Err(err).Msg("error reading part")
			break
		}

		// Reports do not always arrive as a formal attachment: Fastmail
		// sends its aggregate reports with Content-Disposition: inline,
		// which go-message surfaces as *mail.InlineHeader (it uses the
		// same type for text/* parts with no disposition at all). Both
		// types wrap a message.Header, and AttachmentHeader.Filename
		// already reads the disposition filename with a Content-Type name
		// fallback, so wrap the inline header and leave the decision to
		// isDMARCAttachment below.
		var h *mail.AttachmentHeader
		inline := false
		switch ph := part.Header.(type) {
		case *mail.AttachmentHeader:
			h = ph
		case *mail.InlineHeader:
			h = &mail.AttachmentHeader{Header: ph.Header}
			inline = true
		default:
			continue
		}

		filename, _ := h.Filename()
		contentType, _, _ := h.ContentType()

		// An unnamed inline text part is the human-readable body of the
		// mail, not a report. Skip it before reading so a large HTML body
		// is neither buffered nor content-sniffed: a body that merely
		// quotes "<feedback" must not turn into an attachment.
		if inline && filename == "" && strings.HasPrefix(strings.ToLower(contentType), "text/") {
			continue
		}

		data, err := io.ReadAll(part.Body)
		if err != nil {
			c.log.Warn().Err(err).Msg("error reading attachment")
			continue
		}

		if isEmbeddedMessage(contentType, filename, data) {
			attachments = append(attachments, c.collectAttachments(bytes.NewReader(data), depth+1)...)
			continue
		}

		// Outlook's binary .msg (OLE compound file) needs a CFB reader;
		// log it instead of silently dropping. Add a parser if it shows up often.
		if bytes.HasPrefix(data, []byte{0xd0, 0xcf, 0x11, 0xe0}) {
			c.log.Warn().Str("filename", filename).Msg("unsupported Outlook .msg (OLE) attachment; ask the sender to forward as .eml")
			continue
		}

		if isDMARCAttachment(filename, data) {
			attachments = append(attachments, Attachment{Filename: filename, Data: data})
		}
	}

	return attachments
}

// isEmbeddedMessage reports whether the attachment is itself an email message.
func isEmbeddedMessage(contentType, filename string, data []byte) bool {
	if strings.HasPrefix(strings.ToLower(contentType), "message/rfc822") {
		return true
	}
	lower := strings.ToLower(filename)
	if !strings.HasSuffix(lower, ".eml") && !strings.HasSuffix(lower, ".msg") {
		return false
	}
	// .msg is ambiguous: rfc822 text from most relays, OLE binary from Outlook.
	return bytes.Contains(headOf(data), []byte(":"))
}

func headOf(data []byte) []byte {
	if len(data) > 512 {
		return data[:512]
	}
	return data
}

// isDMARCAttachment checks if an attachment is likely a DMARC report, by
// filename or — for senders using opaque names — by sniffing the content.
func isDMARCAttachment(filename string, data []byte) bool {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".xml") ||
		strings.HasSuffix(lower, ".gz") ||
		strings.HasSuffix(lower, ".zip") ||
		strings.Contains(lower, "dmarc") {
		return true
	}

	// Content sniff: gzip/zip magic, or a raw <feedback> document.
	return bytes.HasPrefix(data, []byte{0x1f, 0x8b}) ||
		bytes.HasPrefix(data, []byte("PK\x03\x04")) ||
		bytes.Contains(headOf(data), []byte("<feedback"))
}
