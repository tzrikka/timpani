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
func CountWebhookEvent(t time.Time, event string, statusCode int) int {
	muIn.Lock()
	defer muIn.Unlock()

	record := []string{t.Format(time.RFC3339), event, strconv.Itoa(statusCode)}
	writeLineToFile(DefaultMetricsFileIn, record)
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
	writeLineToFile(DefaultMetricsFileOut, record)
}

func writeLineToFile(filename string, record []string) {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //gosec:disable G304 -- filename not controlled by user
	if err != nil {
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	_ = w.Write(record)
	w.Flush()
}
