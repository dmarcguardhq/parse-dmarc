package parser

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var testNow = time.Date(2026, 11, 20, 12, 0, 0, 0, time.UTC)

const rfc9990Report = `<?xml version="1.0" encoding="UTF-8"?>
<feedback xmlns="urn:ietf:params:xml:ns:dmarc-2.0">
  <version>1.0</version>
  <report_metadata>
    <org_name>Mail.Receiver.example</org_name>
    <email>dmarc-reports@mail.receiver.example</email>
    <report_id> 3v98abbp8ya9n3va8yr8oa3ya </report_id>
    <date_range><begin>1763337600</begin><end>1763424000</end></date_range>
    <generator>Example DMARC Aggregate Reporter v1.2</generator>
  </report_metadata>
  <policy_published>
    <domain>Example.COM</domain>
    <p>quarantine</p><sp>none</sp><np>reject</np>
    <adkim>r</adkim><aspf>r</aspf>
    <discovery_method>treewalk</discovery_method>
    <testing>n</testing>
  </policy_published>
  <record>
    <row>
      <source_ip>192.0.2.123</source_ip>
      <count>123</count>
      <policy_evaluated>
        <disposition>Pass</disposition><dkim>PASS</dkim><spf>fail</spf>
        <reason><type>local_policy</type><comment>allowlisted</comment></reason>
      </policy_evaluated>
    </row>
    <identifiers><header_from>Example.com</header_from></identifiers>
    <auth_results>
      <dkim><domain>example.com</domain><selector>abc123</selector><result>pass</result></dkim>
      <spf><domain>Bounce.Example.org</domain><scope>mfrom</scope><result>fail</result><human_result>no match</human_result></spf>
    </auth_results>
  </record>
</feedback>`

func TestParseRFC9990Report(t *testing.T) {
	f, err := parseDMARC([]byte(rfc9990Report), testNow)
	if err != nil {
		t.Fatal(err)
	}
	if f.SchemaVersion() != "2.0" {
		t.Errorf("schema version = %q, want 2.0", f.SchemaVersion())
	}
	if f.ReportMetadata.ReportID != "3v98abbp8ya9n3va8yr8oa3ya" || f.ReportMetadata.Generator == "" {
		t.Errorf("metadata: %+v", f.ReportMetadata)
	}
	pp := f.PolicyPublished
	if pp.Domain != "example.com" || pp.NP != "reject" || pp.DiscoveryMethod != "treewalk" || pp.Testing != "n" {
		t.Errorf("policy: %+v", pp)
	}
	rec := f.Records[0]
	if pe := rec.Row.PolicyEvaluated; pe.Disposition != "pass" || pe.DKIM != "pass" || pe.Reason[0].Type != "local_policy" {
		t.Errorf("policy evaluated not normalized: %+v", pe)
	}
	if rec.Identifiers.HeaderFrom != "example.com" {
		t.Errorf("header_from = %q", rec.Identifiers.HeaderFrom)
	}
	// envelope_from is missing: it is the checked SPF domain.
	if rec.Identifiers.EnvelopeFrom != "bounce.example.org" {
		t.Errorf("envelope_from = %q", rec.Identifiers.EnvelopeFrom)
	}
	if rec.AuthResults.SPF[0].HumanResult != "no match" {
		t.Errorf("SPF human_result = %q", rec.AuthResults.SPF[0].HumanResult)
	}
	if f.GetDMARCCompliantCount() != 123 {
		t.Errorf("compliant = %d, want 123 (upper-case PASS)", f.GetDMARCCompliantCount())
	}
}

func TestParseDMARCLegacySchemaVersion(t *testing.T) {
	doc := strings.Replace(rfc9990Report, ` xmlns="urn:ietf:params:xml:ns:dmarc-2.0"`, "", 1)
	f, err := parseDMARC([]byte(doc), testNow)
	if err != nil {
		t.Fatal(err)
	}
	if f.SchemaVersion() != "1.0" {
		t.Errorf("schema version = %q, want 1.0", f.SchemaVersion())
	}
}

func TestParseDMARCOrgFromEmail(t *testing.T) {
	doc := strings.Replace(rfc9990Report, "<org_name>Mail.Receiver.example</org_name>", "", 1)
	f, err := parseDMARC([]byte(doc), testNow)
	if err != nil {
		t.Fatal(err)
	}
	if f.ReportMetadata.OrgName != "mail.receiver.example" {
		t.Errorf("org_name = %q", f.ReportMetadata.OrgName)
	}
}

func TestParseDMARCRejects(t *testing.T) {
	cases := map[string]string{
		"no report id": strings.Replace(rfc9990Report, "<report_id> 3v98abbp8ya9n3va8yr8oa3ya </report_id>", "", 1),
		"no domain":    strings.Replace(rfc9990Report, "<domain>Example.COM</domain>", "", 1),
		"no reporter": strings.Replace(strings.Replace(rfc9990Report,
			"<org_name>Mail.Receiver.example</org_name>", "", 1),
			"<email>dmarc-reports@mail.receiver.example</email>", "", 1),
		"reversed range":   strings.Replace(rfc9990Report, "<begin>1763337600</begin>", "<begin>1763500000</begin>", 1),
		"span over a year": strings.Replace(rfc9990Report, "<begin>1763337600</begin>", "<begin>1700000000</begin>", 1),
		"future":           strings.Replace(rfc9990Report, "<end>1763424000</end>", "<end>1999999999</end>", 1),
	}
	for name, doc := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseDMARC([]byte(doc), testNow); !errors.Is(err, ErrInvalidDMARC) {
				t.Fatalf("err = %v, want ErrInvalidDMARC", err)
			}
		})
	}
}
