// Package metrics provides functions to record metrics data.
// It is a very thin layer over OpenTelemetry, but it can
// also write logs to local files for simple setups.
package metrics

import (
	"encoding/csv"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const (
	DefaultMetricsFileIn  = "timpani_metrics_in.csv"
	DefaultMetricsFileOut = "timpani_metrics_out.csv"
)

var (
	muIn  sync.Mutex
	muOut sync.Mutex
)

// CountWebhookEvent counts incoming webhook events as a metric. It returns the HTTP
// status code that was passed to it, in order to return it to the remote HTTP client.
func CountWebhookEvent(l zerolog.Logger, t time.Time, event string, statusCode int) int {
	muIn.Lock()
	defer muIn.Unlock()

	record := []string{t.Format(time.RFC3339), event, strconv.Itoa(statusCode)}
	writeLineToFile(&l, DefaultMetricsFileIn, record)
	return statusCode
}

// CountAPICall counts outgoing API calls as a metric.
func CountAPICall(t time.Time, method string, err error) {
	muOut.Lock()
	defer muOut.Unlock()

	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	record := []string{t.Format(time.RFC3339), method, errMsg}
	writeLineToFile(nil, DefaultMetricsFileOut, record)
}

func writeLineToFile(l *zerolog.Logger, filename string, record []string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		if l != nil {
			l.Error().Err(err).Msg("failed to open metrics file")
		}
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(record); err != nil {
		if l != nil {
			l.Error().Err(err).Msg("failed to write metrics file")
		}
	}
	w.Flush()
}
