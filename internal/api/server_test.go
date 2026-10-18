package api

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/dmarcguardhq/parse-dmarc/internal/parser"
	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
)

// TestHandleReportsLimitCapped pins the pagination ceiling the MCP getReports
// tool already applies: a limit above storage.MaxReportLimit returns
// MaxReportLimit rows, not the whole table.
func TestHandleReportsLimitCapped(t *testing.T) {
	store, err := storage.NewStorage(":memory:")
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer func() { _ = store.Close() }()

	total := storage.MaxReportLimit + 50
	for i := range total {
		feedback := &parser.Feedback{
			ReportMetadata: parser.ReportMetadata{
				ReportID:  fmt.Sprintf("report-%d", i),
				OrgName:   "example.com",
				DateRange: parser.DateRange{Begin: int64(i), End: int64(i + 1)},
			},
			PolicyPublished: parser.PolicyPublished{Domain: "example.com", P: "none"},
		}
		if err := store.SaveReport(feedback); err != nil {
			t.Fatalf("Failed to save report %d: %v", i, err)
		}
	}

	log := zerolog.Nop()
	srv := NewServer(store, "127.0.0.1", 0, nil, &log)

	for _, tc := range []struct {
		name string
		url  string
		want int
	}{
		{"no limit uses the default", "/api/reports", 50},
		{"a limit under the ceiling is honoured", "/api/reports?limit=5", 5},
		{"a limit over the ceiling is clamped", fmt.Sprintf("/api/reports?limit=%d", total*1000), storage.MaxReportLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			srv.handleReports(rec, httptest.NewRequest("GET", tc.url, nil))

			if rec.Code != 200 {
				t.Fatalf("status = %d, want 200: %s", rec.Code, rec.Body.String())
			}
			var reports []storage.ReportSummary
			if err := json.Unmarshal(rec.Body.Bytes(), &reports); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if len(reports) != tc.want {
				t.Errorf("got %d reports, want %d", len(reports), tc.want)
			}
		})
	}
}
