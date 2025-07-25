// Package listeners defines standard types for input parameters and function
// signatures which are used by all the handler functions in [pkg/listeners].
//
// [pkg/listeners]: https://pkg.go.dev/github.com/tzrikka/timpani/pkg/listeners
package listeners

import (
	"context"
	"net/http"
	"net/url"
)

type TemporalConfig struct {
	HostPort  string
	Namespace string
	TaskQueue string
}

type RequestData struct {
	PathSuffix  string
	Headers     http.Header
	WebForm     url.Values
	RawPayload  []byte
	JSONPayload map[string]any
	LinkSecrets map[string]string
	Temporal    TemporalConfig
}

type LinkData struct {
	ID       string
	Template string
	Secrets  map[string]string
}

type WebhookHandlerFunc func(ctx context.Context, w http.ResponseWriter, r RequestData) int

type ConnHandlerFunc func(ctx context.Context, data LinkData) error

const (
	WaitForEventWorkflow = "timpani.waitForEvent"
)

type WaitForEventRequest struct {
	Source  string `json:"source"`
	Name    string `json:"name"`
	Timeout string `json:"timeout,omitempty"`
}
