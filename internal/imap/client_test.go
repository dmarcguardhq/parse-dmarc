package imap

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/meysam81/parse-dmarc/internal/config"
	"github.com/rs/zerolog"
)

const feedbackXML = `<?xml version="1.0"?><feedback><report_metadata><report_id>1</report_id></report_metadata></feedback>`

func zipped(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.Create("google.com!example.com!1!2.xml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte(feedbackXML)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func gzipped(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write([]byte(feedbackXML)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// Fastmail sends its aggregate reports as an inline part carrying a filename,
// alongside a plain text body that must not be picked up as a report.
func TestCollectAttachmentsInlineDisposition(t *testing.T) {
	msg := "From: reports@fastmaildmarc.com\r\nMIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=b\r\n\r\n" +
		"--b\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" +
		"This is an aggregate report for example.com.\r\n" +
		"--b\r\nContent-Type: application/gzip\r\n" +
		"Content-Disposition: inline; filename=\"fastmail.com!example.com!1!2.xml.gz\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		encodeB64(gzipped(t)) + "\r\n--b--\r\n"

	log := zerolog.Nop()
	c := &Client{log: &log}
	atts := c.collectAttachments(strings.NewReader(msg), 0)
	if len(atts) != 1 {
		t.Fatalf("want 1 attachment, got %d", len(atts))
	}
	if atts[0].Filename != "fastmail.com!example.com!1!2.xml.gz" {
		t.Fatalf("want filename from the inline part, got %q", atts[0].Filename)
	}
	if !bytes.HasPrefix(atts[0].Data, []byte{0x1f, 0x8b}) {
		t.Fatalf("want gzip payload, got %x", atts[0].Data)
	}
}

// A text/xml report carrying only a Content-Type name parameter and no
// Content-Disposition is an InlineHeader to go-message as well; the name
// fallback must still pick it up.
func TestCollectAttachmentsInlineNameOnly(t *testing.T) {
	msg := "From: reporter@example.com\r\nMIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=b\r\n\r\n" +
		"--b\r\nContent-Type: text/xml; name=\"reporter.example!example.com!1!2.xml\"\r\n\r\n" +
		feedbackXML + "\r\n--b--\r\n"

	log := zerolog.Nop()
	c := &Client{log: &log}
	atts := c.collectAttachments(strings.NewReader(msg), 0)
	if len(atts) != 1 {
		t.Fatalf("want 1 attachment, got %d", len(atts))
	}
	if atts[0].Filename != "reporter.example!example.com!1!2.xml" {
		t.Fatalf("want filename from the Content-Type name parameter, got %q", atts[0].Filename)
	}
}

// An unnamed inline body is never a report, even when it quotes the
// "<feedback" element the content sniff looks for.
func TestCollectAttachmentsInlineBodyIsNotAReport(t *testing.T) {
	msg := "From: someone@example.com\r\nMIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=b\r\n\r\n" +
		"--b\r\nContent-Type: text/html; charset=utf-8\r\n\r\n" +
		"<p>The report's <feedback> element was empty.</p>\r\n--b--\r\n"

	log := zerolog.Nop()
	c := &Client{log: &log}
	if atts := c.collectAttachments(strings.NewReader(msg), 0); len(atts) != 0 {
		t.Fatalf("want 0 attachments from an HTML-only body, got %d", len(atts))
	}
}

// A Google report relayed as a message/rfc822 attachment (the ".msg" case).
func TestCollectAttachmentsNestedMessage(t *testing.T) {
	inner := "From: noreply-dmarc-support@google.com\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=in\r\n\r\n" +
		"--in\r\nContent-Type: application/zip\r\n" +
		"Content-Disposition: attachment; filename=\"report.zip\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		encodeB64(zipped(t)) + "\r\n--in--\r\n"

	outer := "From: relay@example.com\r\nMIME-Version: 1.0\r\n" +
		"Content-Type: multipart/mixed; boundary=out\r\n\r\n" +
		"--out\r\nContent-Type: message/rfc822\r\n" +
		"Content-Disposition: attachment; filename=\"Report domain.msg\"\r\n\r\n" +
		inner + "\r\n--out--\r\n"

	log := zerolog.Nop()
	c := &Client{log: &log}
	atts := c.collectAttachments(strings.NewReader(outer), 0)
	if len(atts) != 1 {
		t.Fatalf("want 1 attachment, got %d", len(atts))
	}
	if !bytes.HasPrefix(atts[0].Data, []byte("PK\x03\x04")) {
		t.Fatalf("want zip payload, got %q", atts[0].Data[:4])
	}
}

func TestIsDMARCAttachmentSniff(t *testing.T) {
	if !isDMARCAttachment("opaque.bin", []byte(feedbackXML)) {
		t.Error("content sniff failed for raw feedback XML")
	}
	if isDMARCAttachment("logo.png", []byte("\x89PNG")) {
		t.Error("PNG should not be treated as a DMARC report")
	}
}

func encodeB64(b []byte) string {
	e := base64.StdEncoding.EncodeToString(b)
	var out []string
	for len(e) > 76 {
		out, e = append(out, e[:76]), e[76:]
	}
	return strings.Join(append(out, e), "\r\n")
}

func testCAPEM(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "internal-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func TestTLSConfig(t *testing.T) {
	log := zerolog.Nop()
	caFile := filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(caFile, testCAPEM(t), 0600); err != nil {
		t.Fatal(err)
	}
	junkFile := filepath.Join(t.TempDir(), "junk.pem")
	if err := os.WriteFile(junkFile, []byte("not a certificate"), 0600); err != nil {
		t.Fatal(err)
	}

	c := &Client{log: &log, config: &config.IMAPConfig{Host: "imap.internal"}}
	cfg, err := c.tlsConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerName != "imap.internal" || cfg.InsecureSkipVerify || cfg.RootCAs != nil {
		t.Fatalf("unexpected default tls config: %+v", cfg)
	}

	c.config = &config.IMAPConfig{Host: "imap.internal", TLSSkipVerify: true}
	if cfg, err = c.tlsConfig(); err != nil || !cfg.InsecureSkipVerify {
		t.Fatalf("skip verify not honored: %v %v", cfg, err)
	}

	c.config = &config.IMAPConfig{Host: "imap.internal", TLSCAFile: caFile}
	if cfg, err = c.tlsConfig(); err != nil || cfg.RootCAs == nil {
		t.Fatalf("CA file not loaded: %v %v", cfg, err)
	}

	c.config = &config.IMAPConfig{Host: "imap.internal", TLSCAFile: junkFile}
	if _, err = c.tlsConfig(); err == nil {
		t.Error("want error for a PEM file with no certificates")
	}

	c.config = &config.IMAPConfig{Host: "imap.internal", TLSCAFile: filepath.Join(t.TempDir(), "missing.pem")}
	if _, err = c.tlsConfig(); err == nil {
		t.Error("want error for a missing CA file")
	}
}

// firstClientBytes returns what the server sees from the client after the
// greeting, which tells implicit TLS from STARTTLS from plaintext.
func firstClientBytes(t *testing.T, cfg *config.IMAPConfig) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()

	got := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			got <- ""
			return
		}
		defer func() { _ = conn.Close() }()
		_, _ = conn.Write([]byte("* OK [CAPABILITY IMAP4rev1 STARTTLS] ready\r\n"))
		_ = conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		buf := make([]byte, 64)
		n, _ := conn.Read(buf)
		got <- string(buf[:n])
	}()

	host, port, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	cfg.Host = host
	cfg.Port, err = strconv.Atoi(port)
	if err != nil {
		t.Fatal(err)
	}

	log := zerolog.Nop()
	c := &Client{log: &log, config: cfg}
	// The server hangs up mid-handshake, so errors here are expected; the
	// assertion is on what the client sent, not on a successful connection.
	imapClient, err := c.dial(ln.Addr().String(), &tls.Config{InsecureSkipVerify: true}) // #nosec G402 -- test
	first := <-got
	if err == nil {
		_ = imapClient.Logout()
	}
	return first
}

func TestDialTransportSelection(t *testing.T) {
	// UseTLS defaults to true, so STARTTLS must take precedence over it.
	if got := firstClientBytes(t, &config.IMAPConfig{UseTLS: true, StartTLS: true}); !strings.Contains(strings.ToUpper(got), "STARTTLS") {
		t.Errorf("starttls: want a STARTTLS command, got %q", got)
	}
	if got := firstClientBytes(t, &config.IMAPConfig{UseTLS: true}); !strings.HasPrefix(got, "\x16") {
		t.Errorf("implicit tls: want a TLS handshake record, got %q", got)
	}
	if got := firstClientBytes(t, &config.IMAPConfig{}); got != "" {
		t.Errorf("plaintext: want no unsolicited client traffic, got %q", got)
	}
}
