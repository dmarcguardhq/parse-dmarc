package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	cases := []struct {
		name, allowed, origin, method string
		wantHeader                    string
		wantStatus                    int
	}{
		{"off by default", "", "https://evil.example", "GET", "", 200},
		{"same origin has no Origin header", "", "", "GET", "", 200},
		{"listed origin", "https://dash.example.com, https://other.example", "https://dash.example.com", "GET", "https://dash.example.com", 200},
		{"listed origin, case and slash", "https://Dash.example.com/", "https://dash.example.com", "GET", "https://dash.example.com", 200},
		{"unlisted origin", "https://dash.example.com", "https://evil.example", "GET", "", 200},
		{"wildcard", "*", "https://anyone.example", "GET", "*", 200},
		{"preflight from a listed origin", "https://dash.example.com", "https://dash.example.com", "OPTIONS", "https://dash.example.com", 204},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Server{cors: parseCORSOrigins(tc.allowed)}
			req := httptest.NewRequest(tc.method, "/api/reports", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			rec := httptest.NewRecorder()
			s.corsMiddleware(ok).ServeHTTP(rec, req)
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.wantHeader {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tc.wantHeader)
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
		})
	}
}
