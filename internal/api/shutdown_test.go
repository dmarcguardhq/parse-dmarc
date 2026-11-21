package api

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/dmarcguardhq/parse-dmarc/internal/storage"
)

// Start returns once the context is cancelled and the server has stopped,
// so the caller may close the store right after it.
func TestStartReturnsAfterShutdown(t *testing.T) {
	store, err := storage.NewStorage(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = store.Close() }()
	log := zerolog.Nop()
	srv := NewServer(store, "127.0.0.1", 0, nil, &log, VersionInfo{}, WithShutdownTimeout(time.Second))

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- srv.Start(ctx) }()
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Start did not return after cancellation")
	}
}
