package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/meysam81/parse-dmarc/internal/api"
	"github.com/meysam81/parse-dmarc/internal/config"
	"github.com/meysam81/parse-dmarc/internal/imap"
	"github.com/meysam81/parse-dmarc/internal/logger"
	"github.com/meysam81/parse-dmarc/internal/maildir"
	"github.com/meysam81/parse-dmarc/internal/metrics"
	"github.com/meysam81/parse-dmarc/internal/parser"
	"github.com/meysam81/parse-dmarc/internal/storage"
)

func run(ctx context.Context, version, commit, date string) error {
	if genConfig {
		if err := config.GenerateSample(configPath); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		log.Info().Str("path", configPath).Msg("sample configuration written")
		return nil
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Reinitialize logger with config-derived level
	log = logger.NewLogger(cfg.LogLevel, !cfg.ColoredLogs)

	// Validate required IMAP configuration when fetching is enabled
	// (not serve-only and not MCP mode)
	if !serveOnly && !mcpMode && mcpHTTPAddr == "" {
		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("configuration error: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0o755); err != nil {
		return fmt.Errorf("create database directory: %w", err)
	}

	store, err := storage.NewStorage(cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer func() { _ = store.Close() }()

	// Handle MCP mode
	if mcpMode || mcpHTTPAddr != "" {
		return runMCPServer(ctx, store, version)
	}

	// Initialize metrics if enabled
	var m *metrics.Metrics
	if metricsEnabled {
		m = metrics.New(version, commit, date)
		log.Info().Msg("prometheus metrics enabled at /metrics")
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := api.NewServer(store, cfg.Server.Host, cfg.Server.Port, m, log)
	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- server.Start(ctx)
	}()

	// Refresh metrics on startup
	server.RefreshMetrics()

	if serveOnly {
		log.Info().Msg("running in serve-only mode")
		select {
		case <-ctx.Done():
			log.Info().Msg("shutting down")
		case err := <-serverErrChan:
			if err != nil {
				return fmt.Errorf("server error: %w", err)
			}
		}
		return nil
	}

	if fetchOnce {
		if err := fetchReports(cfg, store, m); err != nil {
			return fmt.Errorf("failed to fetch reports: %w", err)
		}
		server.RefreshMetrics()
		log.Info().Msg("fetch complete")
		return nil
	}

	log.Info().Int("interval_seconds", fetchInterval).Msg("starting continuous fetch mode")

	if err := fetchReports(cfg, store, m); err != nil {
		log.Error().Err(err).Msg("initial fetch failed")
	}
	server.RefreshMetrics()

	ticker := time.NewTicker(time.Duration(fetchInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := fetchReports(cfg, store, m); err != nil {
				log.Error().Err(err).Msg("fetch failed")
			}
			server.RefreshMetrics()
		case <-ctx.Done():
			log.Info().Msg("shutting down")
			return nil
		case err := <-serverErrChan:
			if err != nil {
				return fmt.Errorf("server error: %w", err)
			}
		}
	}
}

// fetchReports reads one batch of reports from the source the configuration
// names: a Maildir when MAILDIR_PATH is set, IMAP otherwise.
func fetchReports(cfg *config.Config, store *storage.Storage, m *metrics.Metrics) error {
	if cfg.Maildir.Path != "" {
		return fetchMaildirReports(cfg, store, m)
	}
	return fetchIMAPReports(cfg, store, m)
}

func fetchIMAPReports(cfg *config.Config, store *storage.Storage, m *metrics.Metrics) error {
	log.Info().Msg("fetching DMARC reports")

	fetchStart := time.Now()
	if m != nil {
		m.FetchCyclesTotal.Inc()
	}

	// Create IMAP client
	connectStart := time.Now()
	client := imap.NewClient(&cfg.IMAP, log)
	if err := client.Connect(); err != nil {
		if m != nil {
			m.RecordIMAPConnection(false, time.Since(connectStart))
			m.FetchErrors.Inc()
		}
		return fmt.Errorf("connect to IMAP server: %w", err)
	}
	if m != nil {
		m.RecordIMAPConnection(true, time.Since(connectStart))
	}
	defer func() { _ = client.Disconnect() }()

	// Fetch reports
	result, err := client.FetchDMARCReports()
	if err != nil {
		if m != nil {
			m.FetchErrors.Inc()
		}
		return fmt.Errorf("fetch DMARC reports: %w", err)
	}

	if m != nil {
		m.ReportsFetched.Add(float64(len(result.Reports)))
	}

	if len(result.Reports) == 0 {
		log.Info().Msg("no new reports found")
		if m != nil {
			m.RecordFetchDuration(time.Since(fetchStart))
			m.LastFetchTimestamp.SetToCurrentTime()
		}
		return nil
	}

	log.Info().Int("count", len(result.Reports)).Msg("processing reports")
	processed, _ := processReports(result.Reports, store, m)

	// Post-processing: mark as seen and/or move messages
	if len(result.MessageIDs) > 0 {
		if cfg.IMAP.MarkAsSeen {
			if err := client.MarkAsSeen(result.MessageIDs); err != nil {
				log.Error().Err(err).Msg("failed to mark messages as seen")
			} else {
				log.Info().Int("count", len(result.MessageIDs)).Msg("marked messages as seen")
			}
		}
		if cfg.IMAP.ProcessedMailbox != "" {
			if err := client.MoveMessages(result.MessageIDs, cfg.IMAP.ProcessedMailbox); err != nil {
				log.Error().Err(err).Msg("failed to move messages to processed mailbox")
			} else {
				log.Info().Int("count", len(result.MessageIDs)).Str("mailbox", cfg.IMAP.ProcessedMailbox).Msg("moved messages to processed mailbox")
			}
		}
	}

	if m != nil {
		m.RecordFetchDuration(time.Since(fetchStart))
		m.LastFetchTimestamp.SetToCurrentTime()
	}

	log.Info().Int("count", processed).Msg("reports processed")
	return nil
}

// fetchMaildirReports reads the messages a MTA delivered into a local Maildir
// and moves the whole batch to cur/ once it has been through the store.
func fetchMaildirReports(cfg *config.Config, store *storage.Storage, m *metrics.Metrics) error {
	log.Info().Str("path", cfg.Maildir.Path).Msg("fetching DMARC reports from Maildir")

	fetchStart := time.Now()
	if m != nil {
		m.FetchCyclesTotal.Inc()
	}

	client, err := maildir.NewClient(&cfg.Maildir, log)
	if err != nil {
		if m != nil {
			m.FetchErrors.Inc()
		}
		return fmt.Errorf("maildir client: %w", err)
	}
	result, err := client.FetchDMARCReports()
	if err != nil {
		if m != nil {
			m.FetchErrors.Inc()
		}
		return fmt.Errorf("fetch DMARC reports from maildir: %w", err)
	}

	if m != nil {
		m.ReportsFetched.Add(float64(len(result.Reports)))
	}

	processed := 0
	var failed []uint32
	if len(result.Reports) > 0 {
		log.Info().Int("count", len(result.Reports)).Msg("processing reports")
		processed, failed = processReports(result.Reports, store, m)
	} else {
		log.Info().Msg("no new reports found")
	}

	// Messages that carried no report at all, and messages whose report did
	// not parse or store, go to the deadletter mailbox: the operator sees
	// them there instead of them passing as processed.
	deadletter := append(result.DeadletterIDs, failed...)
	if len(deadletter) > 0 {
		if err := client.MoveToDeadletter(deadletter); err != nil {
			log.Error().Err(err).Str("mailbox", cfg.Maildir.DeadletterMailbox).Msg("failed to move messages to deadletter mailbox")
		} else {
			log.Info().Int("count", len(deadletter)).Str("mailbox", cfg.Maildir.DeadletterMailbox).Msg("moved messages to deadletter mailbox")
		}
	}

	// Every other message read in this cycle leaves new/: a Maildir has no
	// unseen flag, so a message left there is read again on every following
	// cycle.
	remaining := slices.DeleteFunc(slices.Clone(result.MessageIDs), func(id uint32) bool {
		return slices.Contains(deadletter, id)
	})
	if len(remaining) > 0 {
		if err := client.MoveToCur(remaining); err != nil {
			log.Error().Err(err).Msg("failed to move messages to cur")
		} else {
			log.Info().Int("count", len(remaining)).Msg("moved messages to cur")
		}
	}

	if m != nil {
		m.RecordFetchDuration(time.Since(fetchStart))
		m.LastFetchTimestamp.SetToCurrentTime()
	}

	log.Info().Int("count", processed).Msg("reports processed")
	return nil
}

// processReports parses and stores the report attachments of one batch. It
// returns how many reports were saved, and the IDs of the messages that
// reached the store with nothing, so the caller can flag them to the operator.
func processReports(reports []imap.Report, store *storage.Storage, m *metrics.Metrics) (int, []uint32) {
	processed := 0
	var failed []uint32
	for _, report := range reports {
		before := processed

		for _, attachment := range report.Attachments {
			if m != nil {
				m.AttachmentsTotal.Inc()
			}

			feedback, err := parser.ParseReport(attachment.Data)
			if err != nil {
				log.Warn().Err(err).
					Str("filename", attachment.Filename).
					Str("message_id", report.MessageID).
					Str("from", report.From).
					Msg("failed to parse report")
				if m != nil {
					m.ReportParseErrors.Inc()
				}
				continue
			}
			if m != nil {
				m.ReportsParsed.Inc()
			}

			if err := store.SaveReport(feedback); err != nil {
				log.Error().Err(err).
					Str("report_id", feedback.ReportMetadata.ReportID).
					Str("message_id", report.MessageID).
					Msg("failed to save report")
				if m != nil {
					m.ReportStoreErrors.Inc()
				}
				continue
			}
			if m != nil {
				m.ReportsStored.Inc()
			}

			log.Info().
				Str("report_id", feedback.ReportMetadata.ReportID).
				Str("message_id", report.MessageID).
				Str("org", feedback.ReportMetadata.OrgName).
				Str("domain", feedback.PolicyPublished.Domain).
				Int("messages", feedback.GetTotalMessages()).
				Msg("saved report")
			processed++
		}

		// Nothing in this message reached the store: no report attachment that
		// parsed, or a save that failed.
		if processed == before {
			failed = append(failed, report.UID)
		}
	}
	return processed, failed
}
