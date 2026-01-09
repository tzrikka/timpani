// Package metrics provides functions to record metrics data.
// It is a very thin layer over OpenTelemetry, but it can
// also write logs to local files for simple setups.
package metrics

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/tzrikka/xdg"
)

const (
	DefaultMetricsFileIn  = "metrics/timpani_in_%s.csv"
	DefaultMetricsFileOut = "metrics/timpani_out_%s.csv"

	fileFlags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	filePerms = xdg.NewFilePermissions
)

var (
	muIn  sync.Mutex
	muOut sync.Mutex
)

// IncrementWebhookEventCounter monitors incoming webhook events. It returns the HTTP
// status code that was passed to it, in order to return it to the remote HTTP client.
func IncrementWebhookEventCounter(l *slog.Logger, t time.Time, event string, statusCode int) int {
	muIn.Lock()
	defer muIn.Unlock()

	record := []string{t.Format(time.RFC3339), event, strconv.Itoa(statusCode)}
	if err := appendToCSVFile(DefaultMetricsFileIn, t, record); err != nil {
		l.Error("metrics error: failed to increment signal counter", slog.Any("error", err),
			slog.String("event", event), slog.Int("status", statusCode))
	}

	return statusCode
}

// IncrementAPICallCounter monitors outgoing API calls.
func IncrementAPICallCounter(t time.Time, method string, err error) {
	muOut.Lock()
	defer muOut.Unlock()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	_ = appendToCSVFile(DefaultMetricsFileOut, t, []string{t.Format(time.RFC3339), method, errMsg})
}

func appendToCSVFile(filename string, t time.Time, record []string) error {
	filename = fmt.Sprintf(filename, t.Format(time.DateOnly))
	f, err := os.OpenFile(filename, fileFlags, filePerms) //gosec:disable G304 // Hardcoded path.
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(record); err != nil {
		return err
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}

	return nil
}
