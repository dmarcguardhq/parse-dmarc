package api

import (
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
)

func TestHealthEndpoints(t *testing.T) {
	store, err := storage.NewStorage(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	log := zerolog.Nop()
	srv := NewServer(store, "127.0.0.1", 0, nil, &log, VersionInfo{})

	check := func(name string, handle func(*httptest.ResponseRecorder), want int) {
		t.Helper()
		rec := httptest.NewRecorder()
		handle(rec)
		if rec.Code != want {
			t.Errorf("%s: status %d, want %d", name, rec.Code, want)
		}
	}
	live := func(rec *httptest.ResponseRecorder) {
		srv.handleHealthz(rec, httptest.NewRequest("GET", "/healthz", nil))
	}
	ready := func(rec *httptest.ResponseRecorder) {
		srv.handleReadyz(rec, httptest.NewRequest("GET", "/readyz", nil))
	}

	check("live", live, 200)
	check("ready", ready, 200)

	// A closed database is not ready, and the process is still live.
	_ = store.Close()
	check("ready with a closed database", ready, 503)
	check("live with a closed database", live, 200)
}
