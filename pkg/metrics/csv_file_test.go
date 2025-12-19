package metrics_test

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/tzrikka/timpani/pkg/metrics"
)

func TestIncrementWebhookEventCounter(t *testing.T) {
	t.Chdir(t.TempDir())
	now := time.Now().UTC()

	if err := os.Mkdir("metrics", 0o700); err != nil {
		t.Fatal(err)
	}

	want1 := 200
	got1 := metrics.IncrementWebhookEventCounter(slog.Default(), now, "event", want1)
	if got1 != want1 {
		t.Errorf("IncrementWebhookEventCounter() = %v, want %v", got1, want1)
	}

	f, err := os.ReadFile(fmt.Sprintf(metrics.DefaultMetricsFileIn, now.Format(time.DateOnly)))
	if err != nil {
		t.Fatal(err)
	}

	got2 := string(f)
	want2 := now.Format(time.RFC3339) + ",event,200\n"
	if got2 != want2 {
		t.Errorf("file content = %q, want %q", got2, want2)
	}
}

func TestIncrementAPICallCounter(t *testing.T) {
	t.Chdir(t.TempDir())
	now := time.Now().UTC()

	if err := os.Mkdir("metrics", 0o700); err != nil {
		t.Fatal(err)
	}

	metrics.IncrementAPICallCounter(now, "method 1", nil)
	metrics.IncrementAPICallCounter(now, "method 2", errors.New("some error"))

	f, err := os.ReadFile(fmt.Sprintf(metrics.DefaultMetricsFileOut, now.Format(time.DateOnly)))
	if err != nil {
		t.Fatal(err)
	}

	got := string(f)
	ts := now.Format(time.RFC3339)
	want := fmt.Sprintf("%s,method 1,\n%s,method 2,some error\n", ts, ts)
	if got != want {
		t.Errorf("file content = %q, want %q", got, want)
	}
}
