package parser

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"hash/crc32"
	"runtime"
	"testing"
)

const sampleReport = `<?xml version="1.0"?>
<feedback>
  <version>1.0</version>
  <report_metadata>
    <org_name>example.com</org_name>
    <report_id>abc123</report_id>
  </report_metadata>
  <policy_published><domain>example.com</domain><p>none</p></policy_published>
  <record></record>
</feedback>`

func gzipBytes(t *testing.T, payload []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if _, err := zw.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// zipBytes archives payload as report.xml, declaring declaredSize as its
// uncompressed size. Passing less than len(payload) simulates a bomb whose
// header understates its size to slip past the declared-size pre-check.
func zipBytes(t *testing.T, payload []byte, declaredSize int) []byte {
	t.Helper()
	var deflated bytes.Buffer
	fw, _ := flate.NewWriter(&deflated, flate.BestSpeed)
	if _, err := fw.Write(payload); err != nil {
		t.Fatal(err)
	}
	if err := fw.Close(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, err := zw.CreateRaw(&zip.FileHeader{
		Name:               "report.xml",
		Method:             zip.Deflate,
		CRC32:              crc32.ChecksumIEEE(payload),
		CompressedSize64:   uint64(deflated.Len()),
		UncompressedSize64: uint64(declaredSize),
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(deflated.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// Reports arrive from untrusted senders. Anything over the cap, whether raw or
// inflated from a tiny gzip/zip member, must be rejected and must not allocate
// anywhere near the inflated size.
func TestOversizedReportsAreRejected(t *testing.T) {
	bomb := make([]byte, 16*maxDecompressedSize)
	cases := map[string]struct {
		data    []byte
		wantErr error
	}{
		"raw":       {make([]byte, maxDecompressedSize+1), errReportTooLarge},
		"gzip bomb": {gzipBytes(t, bomb), errReportTooLarge},
		"zip bomb":  {zipBytes(t, bomb, len(bomb)), errReportTooLarge},
		// archive/zip refuses to read past a member's declared size, so an
		// understated header fails as a malformed archive rather than a bomb.
		"zip bomb, lying header": {zipBytes(t, bomb, 100), zip.ErrFormat},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			_, err := ParseReport(tc.data)
			runtime.ReadMemStats(&after)
			grewMB := (int64(after.TotalAlloc) - int64(before.TotalAlloc)) >> 20
			t.Logf("input %d bytes, err=%v, allocated ~%d MB", len(tc.data), err, grewMB)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("want %v, got %v", tc.wantErr, err)
			}
			// Reading stops at the cap, so allocation is a few times the cap
			// (buffer doubling; -race roughly doubles it again). Buffering the
			// whole stream before checking would cost at least the inflated size.
			if bombMB := int64(len(bomb) >> 20); grewMB >= bombMB {
				t.Fatalf("allocated ~%d MB for a %d MB bomb; cap is not bounding the read", grewMB, bombMB)
			}
		})
	}
}

// Every accepted container must still yield the parsed report.
func TestCompressedReportsStillParse(t *testing.T) {
	report := []byte(sampleReport)
	cases := map[string][]byte{
		"raw":  report,
		"gzip": gzipBytes(t, report),
		"zip":  zipBytes(t, report, len(report)),
		// archive/zip locates the directory from the end, so a prefixed archive
		// (self-extracting stub) is a valid zip.
		"zip with leading stub": append([]byte("#!/bin/sh\nstub\n"), zipBytes(t, report, len(report))...),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			fb, err := ParseReport(data)
			if err != nil {
				t.Fatalf("valid report should parse, got %v", err)
			}
			if fb.ReportMetadata.OrgName != "example.com" {
				t.Fatalf("unexpected org name: %q", fb.ReportMetadata.OrgName)
			}
		})
	}
}
