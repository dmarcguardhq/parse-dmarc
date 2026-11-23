package parser

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// SchemaNamespace2 is the XML namespace of RFC 9990 aggregate reports. RFC
// 7489 reports carry no namespace.
const SchemaNamespace2 = "urn:ietf:params:xml:ns:dmarc-2.0"

// Feedback represents the root DMARC aggregate report structure (RFC 7489,
// RFC 9990). The element names match in both schemas; XMLName.Space tells
// them apart.
type Feedback struct {
	XMLName         xml.Name        `xml:"feedback"`
	Version         string          `xml:"version"`
	ReportMetadata  ReportMetadata  `xml:"report_metadata"`
	PolicyPublished PolicyPublished `xml:"policy_published"`
	Records         []Record        `xml:"record"`
}

// ReportMetadata contains information about the report
type ReportMetadata struct {
	OrgName          string    `xml:"org_name"`
	Email            string    `xml:"email"`
	ExtraContactInfo string    `xml:"extra_contact_info,omitempty"`
	ReportID         string    `xml:"report_id"`
	DateRange        DateRange `xml:"date_range"`
	Errors           []string  `xml:"error,omitempty"`
	// Generator names the software that wrote the report (RFC 9990).
	Generator string `xml:"generator,omitempty"`
}

// DateRange specifies the time range for the report
type DateRange struct {
	Begin int64 `xml:"begin"`
	End   int64 `xml:"end"`
}

// PolicyPublished represents the DMARC policy as published in DNS
type PolicyPublished struct {
	Domain string `xml:"domain"`
	ADKIM  string `xml:"adkim,omitempty"` // DKIM alignment mode (r=relaxed, s=strict)
	ASPF   string `xml:"aspf,omitempty"`  // SPF alignment mode (r=relaxed, s=strict)
	P      string `xml:"p"`               // Policy (none, quarantine, reject)
	SP     string `xml:"sp,omitempty"`    // Subdomain policy
	PCT    int    `xml:"pct,omitempty"`   // Percentage of messages to filter (RFC 7489 only)
	FO     string `xml:"fo,omitempty"`    // Failure reporting options
	// RFC 9990 additions.
	NP              string `xml:"np,omitempty"`               // Non-existent subdomain policy
	DiscoveryMethod string `xml:"discovery_method,omitempty"` // psl or treewalk
	Testing         string `xml:"testing,omitempty"`          // y: the policy is in test mode
}

// Record represents a single record in the aggregate report
type Record struct {
	Row         Row         `xml:"row"`
	Identifiers Identifiers `xml:"identifiers"`
	AuthResults AuthResults `xml:"auth_results"`
}

// Row contains policy evaluation results
type Row struct {
	SourceIP        string          `xml:"source_ip"`
	Count           int             `xml:"count"`
	PolicyEvaluated PolicyEvaluated `xml:"policy_evaluated"`
}

// PolicyEvaluated shows the result of policy evaluation
type PolicyEvaluated struct {
	Disposition string   `xml:"disposition"` // none, quarantine, reject
	DKIM        string   `xml:"dkim"`        // pass, fail
	SPF         string   `xml:"spf"`         // pass, fail
	Reason      []Reason `xml:"reason,omitempty"`
}

// Reason explains policy override
type Reason struct {
	Type    string `xml:"type"`
	Comment string `xml:"comment,omitempty"`
}

// Identifiers contains message identifiers
type Identifiers struct {
	EnvelopeTo   string `xml:"envelope_to,omitempty"`
	EnvelopeFrom string `xml:"envelope_from,omitempty"`
	HeaderFrom   string `xml:"header_from"`
}

// AuthResults contains authentication results
type AuthResults struct {
	DKIM []DKIMResult `xml:"dkim,omitempty"`
	SPF  []SPFResult  `xml:"spf"`
}

// DKIMResult represents DKIM authentication result
type DKIMResult struct {
	Domain      string `xml:"domain"`
	Selector    string `xml:"selector,omitempty"`
	Result      string `xml:"result"` // none, pass, fail, policy, neutral, temperror, permerror
	HumanResult string `xml:"human_result,omitempty"`
}

// SPFResult represents SPF authentication result
type SPFResult struct {
	Domain      string `xml:"domain"`
	Scope       string `xml:"scope,omitempty"` // helo, mfrom
	Result      string `xml:"result"`          // none, neutral, pass, fail, softfail, temperror, permerror
	HumanResult string `xml:"human_result,omitempty"`
}

// ParseReport parses a DMARC aggregate report from raw data: XML, or XML in
// a gzip or zip container.
func ParseReport(data []byte) (*Feedback, error) {
	return parseDMARC(data, time.Now())
}

func parseDMARC(data []byte, now time.Time) (*Feedback, error) {
	decompressed, err := decompress(data)
	if err != nil {
		return nil, err
	}
	return parseFeedbackXML(decompressed, now)
}

func parseFeedbackXML(doc []byte, now time.Time) (*Feedback, error) {
	var feedback Feedback
	if err := xml.Unmarshal(doc, &feedback); err != nil {
		return nil, fmt.Errorf("XML parsing failed: %w", err)
	}
	feedback.normalize()
	if err := feedback.validate(now); err != nil {
		return nil, err
	}
	return &feedback, nil
}

var utf8BOM = []byte{0xef, 0xbb, 0xbf}

// decompress returns the report document inside data, bounded by
// maxDecompressedSize.
func decompress(data []byte) ([]byte, error) {
	// A raw (uncompressed) report gets the same ceiling as a decompressed one.
	if len(data) > maxDecompressedSize {
		return nil, errReportTooLarge
	}
	decompressed, err := tryDecompress(data)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}
	// Neither encoding/xml nor encoding/json accepts a byte order mark,
	// which some reporters write.
	return bytes.TrimPrefix(decompressed, utf8BOM), nil
}

// SchemaVersion is "2.0" for an RFC 9990 report and the report's own
// <version> (or "draft" when it has none) otherwise.
func (f *Feedback) SchemaVersion() string {
	if f.XMLName.Space == SchemaNamespace2 {
		return "2.0"
	}
	if f.Version != "" {
		return f.Version
	}
	return "draft"
}

// normalize evens out what reporters disagree on, so the same outcome is
// stored and counted the same way: result words in lower case, a missing
// org_name taken from the reporter's address, a missing envelope_from taken
// from the SPF domain that was checked.
func (f *Feedback) normalize() {
	md := &f.ReportMetadata
	md.OrgName = strings.TrimSpace(md.OrgName)
	md.ReportID = strings.Trim(strings.TrimSpace(md.ReportID), "<>")
	if md.OrgName == "" {
		if _, domain, ok := strings.Cut(md.Email, "@"); ok {
			md.OrgName = strings.ToLower(strings.TrimSpace(domain))
		}
	}

	pp := &f.PolicyPublished
	pp.Domain = strings.ToLower(strings.TrimSpace(pp.Domain))
	pp.P = lowerTrim(pp.P)
	pp.SP = lowerTrim(pp.SP)
	pp.NP = lowerTrim(pp.NP)
	pp.ADKIM = lowerTrim(pp.ADKIM)
	pp.ASPF = lowerTrim(pp.ASPF)
	pp.DiscoveryMethod = lowerTrim(pp.DiscoveryMethod)
	pp.Testing = lowerTrim(pp.Testing)

	for i := range f.Records {
		r := &f.Records[i]
		r.Row.SourceIP = strings.TrimSpace(r.Row.SourceIP)
		pe := &r.Row.PolicyEvaluated
		pe.Disposition = lowerTrim(pe.Disposition)
		pe.DKIM = lowerTrim(pe.DKIM)
		pe.SPF = lowerTrim(pe.SPF)
		for j := range pe.Reason {
			pe.Reason[j].Type = lowerTrim(pe.Reason[j].Type)
		}

		id := &r.Identifiers
		id.HeaderFrom = strings.ToLower(strings.TrimSpace(id.HeaderFrom))
		id.EnvelopeFrom = strings.ToLower(strings.TrimSpace(id.EnvelopeFrom))
		id.EnvelopeTo = strings.ToLower(strings.TrimSpace(id.EnvelopeTo))

		for j := range r.AuthResults.DKIM {
			d := &r.AuthResults.DKIM[j]
			d.Domain = strings.ToLower(strings.TrimSpace(d.Domain))
			d.Result = lowerTrim(d.Result)
		}
		for j := range r.AuthResults.SPF {
			sr := &r.AuthResults.SPF[j]
			sr.Domain = strings.ToLower(strings.TrimSpace(sr.Domain))
			sr.Scope = lowerTrim(sr.Scope)
			sr.Result = lowerTrim(sr.Result)
		}
		if id.EnvelopeFrom == "" && len(r.AuthResults.SPF) > 0 {
			id.EnvelopeFrom = r.AuthResults.SPF[len(r.AuthResults.SPF)-1].Domain
		}
	}
}

func lowerTrim(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// Bounds on a report's date range: RFC 9990 reports cover at most a day,
// and a span beyond a year or a date in the future is a broken or forged
// report that would distort every chart it lands in.
const (
	maxReportSpan = 366 * 24 * time.Hour
	maxFutureSkew = 48 * time.Hour
)

// ErrInvalidDMARC wraps every reason a well-formed XML document is still not
// a usable aggregate report.
var ErrInvalidDMARC = errors.New("invalid DMARC aggregate report")

func (f *Feedback) validate(now time.Time) error {
	md := f.ReportMetadata
	switch {
	case md.OrgName == "":
		return fmt.Errorf("%w: no org_name or reporter email", ErrInvalidDMARC)
	case md.ReportID == "":
		return fmt.Errorf("%w: no report_id", ErrInvalidDMARC)
	case f.PolicyPublished.Domain == "":
		return fmt.Errorf("%w: no policy_published domain", ErrInvalidDMARC)
	}
	if err := validateDateRange(time.Unix(md.DateRange.Begin, 0), time.Unix(md.DateRange.End, 0), now); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidDMARC, err)
	}
	return nil
}

func validateDateRange(begin, end, now time.Time) error {
	switch {
	case end.Before(begin):
		return errors.New("date range ends before it begins")
	case end.Sub(begin) > maxReportSpan:
		return fmt.Errorf("date range spans more than %d days", int(maxReportSpan.Hours()/24))
	case begin.After(now.Add(maxFutureSkew)), end.After(now.Add(maxFutureSkew)):
		return errors.New("date range is in the future")
	}
	return nil
}

// tryDecompress attempts to decompress data (gzip or zip). Detection is
// decisive so that a valid-but-oversized member is reported as an error
// instead of silently falling through to "return the raw bytes" (which would
// hand a bomb to the XML parser and re-buffer it).
func tryDecompress(data []byte) ([]byte, error) {
	// gzip mandates its magic bytes, so this check is exact.
	if bytes.HasPrefix(data, []byte{0x1f, 0x8b}) {
		return decompressGzip(data)
	}

	// Let archive/zip decide: it locates the central directory from the end,
	// so archives with a leading stub (self-extracting) are recognised too,
	// which a fixed "PK\x03\x04" prefix check would miss.
	if zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data))); err == nil {
		return decompressZip(zipReader)
	}

	// Not compressed.
	return data, nil
}

// maxDecompressedSize bounds a single report, raw or decompressed. DMARC
// aggregate reports are small (usually well under a megabyte); a generous
// 16 MB ceiling stops a malicious sender from turning a tiny gzip/zip member
// into a multi-gigabyte allocation (a decompression bomb).
const maxDecompressedSize = 16 * 1024 * 1024

var errReportTooLarge = fmt.Errorf("report exceeds %d bytes", maxDecompressedSize)

// readAllLimited reads from r but rejects streams that exceed
// maxDecompressedSize instead of allocating without bound. The buffer starts
// small so normal reports don't pay for the cap; growth stops at the cap, so a
// bomb costs at most a few times maxDecompressedSize in transient allocation.
func readAllLimited(r io.Reader) ([]byte, error) {
	// Read one byte past the cap so we can tell "exactly at the cap" from
	// "over the cap".
	buf := bytes.NewBuffer(make([]byte, 0, 64*1024))
	n, err := io.Copy(buf, io.LimitReader(r, maxDecompressedSize+1))
	if err != nil {
		return nil, err
	}
	if n > maxDecompressedSize {
		return nil, errReportTooLarge
	}
	return buf.Bytes(), nil
}

// decompressGzip decompresses gzip data
func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	return readAllLimited(reader)
}

// decompressZip returns the first XML member of the archive.
func decompressZip(zipReader *zip.Reader) ([]byte, error) {
	// Pick the first XML member; archives may carry directories or __MACOSX noise.
	for _, file := range zipReader.File {
		if file.FileInfo().IsDir() || !strings.HasSuffix(strings.ToLower(file.Name), ".xml") {
			continue
		}
		// Reject a member whose declared uncompressed size is already over the
		// cap before opening it. archive/zip stops a read that exceeds the
		// declared size with ErrFormat, so an understated header can't bypass this.
		if file.UncompressedSize64 > maxDecompressedSize {
			return nil, errReportTooLarge
		}
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer func() { _ = rc.Close() }()

		return readAllLimited(rc)
	}

	return nil, fmt.Errorf("no XML file in zip archive")
}

// GetDateRange returns the date range as time.Time objects
func (f *Feedback) GetDateRange() (time.Time, time.Time) {
	begin := time.Unix(f.ReportMetadata.DateRange.Begin, 0)
	end := time.Unix(f.ReportMetadata.DateRange.End, 0)
	return begin, end
}

// GetTotalMessages returns the total count of messages in the report
func (f *Feedback) GetTotalMessages() int {
	total := 0
	for _, record := range f.Records {
		total += record.Row.Count
	}
	return total
}

// GetDMARCCompliantCount returns count of DMARC-compliant messages
func (f *Feedback) GetDMARCCompliantCount() int {
	count := 0
	for _, record := range f.Records {
		if record.Row.PolicyEvaluated.DKIM == "pass" || record.Row.PolicyEvaluated.SPF == "pass" {
			count += record.Row.Count
		}
	}
	return count
}

// NormalizeForJSON ensures all slice fields are initialized (not nil) to produce
// valid JSON that matches the MCP output schema. The MCP SDK infers JSON schemas
// from Go types, and nil slices serialize as null which violates the array type
// requirement in the schema.
func (f *Feedback) NormalizeForJSON() {
	if f == nil {
		return
	}

	// Normalize ReportMetadata.Errors
	if f.ReportMetadata.Errors == nil {
		f.ReportMetadata.Errors = []string{}
	}

	// Normalize Records slice
	if f.Records == nil {
		f.Records = []Record{}
	}

	// Normalize each record's nested slices
	for i := range f.Records {
		// Normalize PolicyEvaluated.Reason
		if f.Records[i].Row.PolicyEvaluated.Reason == nil {
			f.Records[i].Row.PolicyEvaluated.Reason = []Reason{}
		}

		// Normalize AuthResults.DKIM
		if f.Records[i].AuthResults.DKIM == nil {
			f.Records[i].AuthResults.DKIM = []DKIMResult{}
		}

		// Normalize AuthResults.SPF
		if f.Records[i].AuthResults.SPF == nil {
			f.Records[i].AuthResults.SPF = []SPFResult{}
		}
	}
}
