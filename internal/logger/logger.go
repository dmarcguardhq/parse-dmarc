package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Log formats.
const (
	// FormatConsole is human-readable text, the default.
	FormatConsole = "console"
	// FormatJSON is one JSON object per line, for log collectors.
	FormatJSON = "json"
)

// NewLogger returns a console logger at the given level.
func NewLogger(logLevel string, noColor bool) *zerolog.Logger {
	l, _ := New(logLevel, FormatConsole, noColor)
	return l
}

// New returns a logger at the given level writing to stderr in the given
// format. An unknown format returns a console logger and an error.
func New(logLevel, format string, noColor bool) (*zerolog.Logger, error) {
	return newLogger(os.Stderr, logLevel, format, noColor)
}

func newLogger(out io.Writer, logLevel, format string, noColor bool) (*zerolog.Logger, error) {
	zerolog.TimeFieldFormat = time.RFC3339

	level := zerolog.InfoLevel
	switch strings.ToLower(logLevel) {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	case "critical":
		level = zerolog.FatalLevel
	}

	var err error
	switch strings.ToLower(format) {
	case FormatJSON:
	case FormatConsole, "":
		out = zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339, NoColor: noColor}
	default:
		err = fmt.Errorf("unknown log format %q: want %s or %s", format, FormatConsole, FormatJSON)
		out = zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339, NoColor: noColor}
	}

	l := zerolog.New(out).Level(level).With().Caller().Timestamp().Logger()
	return &l, err
}
