package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"

	"github.com/meysam81/parse-dmarc/internal/parser"
	"github.com/meysam81/parse-dmarc/internal/storage"
)

// The reports endpoint must not hand out more rows than the MCP tool does: one
// unauthenticated request should not buffer the whole table.
func TestHandleReportsCapsLimit(t *testing.T) {
	store, err := storage.NewStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer func() { _ = store.Close() }()

	for i := 0; i < maxAPILimit+20; i++ {
		feedback := &parser.Feedback{
			Version: "1.0",
			ReportMetadata: parser.ReportMetadata{
				OrgName:  "example-reporter.tld",
				Email:    "dmarc@example-reporter.tld",
				ReportID: fmt.Sprintf("report-%d", i),
				DateRange: parser.DateRange{
					Begin: 1780610400,
					End:   1780696800,
				},
			},
			PolicyPublished: parser.PolicyPublished{
				Domain: "example.com",
				P:      "none",
			},
			Records: []parser.Record{{
				Row: parser.Row{
					SourceIP: "192.0.2.1",
					Count:    1,
					PolicyEvaluated: parser.PolicyEvaluated{
						Disposition: "none",
						DKIM:        "pass",
						SPF:         "pass",
					},
				},
				Identifiers: parser.Identifiers{HeaderFrom: "example.com"},
			}},
		}
		if err := store.SaveReport(feedback); err != nil {
			t.Fatalf("Failed to save report: %v", err)
		}
	}

	logger := zerolog.Nop()
	server := &Server{storage: store, log: &logger}

	tests := []struct {
		name  string
		query string
		want  int
	}{
		{"capped", "?limit=1000", maxAPILimit},
		{"smaller limit honoured", "?limit=5", 5},
		{"default", "", 50},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			server.handleReports(recorder, httptest.NewRequest(http.MethodGet, "/api/reports"+test.query, nil))

			if recorder.Code != http.StatusOK {
				t.Fatalf("unexpected status %d: %s", recorder.Code, recorder.Body.String())
			}

			var reports []map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &reports); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if len(reports) != test.want {
				t.Fatalf("expected %d reports, got %d", test.want, len(reports))
			}
		})
	}
}
