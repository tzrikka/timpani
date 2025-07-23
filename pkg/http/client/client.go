// Package client provides a simple, generic HTTP client
// for sending GET and POST requests to external APIs,
// which is used by other packages under [pkg/api].
//
// [pkg/api]: https://pkg.go.dev/github.com/tzrikka/timpani/pkg/api
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.temporal.io/sdk/temporal"
)

const (
	timeout = 3 * time.Second
	maxSize = 10 << 20 // 10 MiB.
)

// HTTPRequest sends an HTTP GET or POST request to an external API.
// For GET requests, the queryOrJSONBody parameter is expected to be
// [url.Values]. For POST requests, it should be any struct that can be
// encoded as JSON. Some errors (failure to construct a request or decode
// a response body) are returned as non-retryable [temporal.ApplicationError]s.
//
// [temporal.ApplicationError]: https://pkg.go.dev/go.temporal.io/temporal#ApplicationError
func HTTPRequest(ctx context.Context, httpMethod, u, authToken string, queryOrJSONBody any) ([]byte, error) {
	req, cancel, err := constructRequest(ctx, httpMethod, u, authToken, queryOrJSONBody)
	if err != nil {
		return nil, err
	}
	defer cancel()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		return nil, fmt.Errorf("failed to read HTTP response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		msg := resp.Status
		if len(body) > 0 {
			msg = fmt.Sprintf("%s: %s", msg, string(body))
		}
		return nil, errors.New(msg)
	}

	return body, nil
}

func constructRequest(ctx context.Context, method, u, token string, queryOrJSONBody any) (*http.Request, context.CancelFunc, error) {
	if method == http.MethodGet {
		u = fmt.Sprintf("%s?%s", u, queryOrJSONBody.(url.Values).Encode())
	}

	b, err := requestBody(method, queryOrJSONBody)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	req, err := http.NewRequestWithContext(ctx, method, u, b)
	if err != nil {
		cancel()
		msg := "failed to construct HTTP request: " + err.Error()
		return nil, nil, temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	return req, cancel, nil
}

func requestBody(method string, queryOrJSONBody any) (io.Reader, error) {
	if method == http.MethodGet {
		return http.NoBody, nil
	}

	// HTTP POST.
	b, err := json.Marshal(queryOrJSONBody)
	if err != nil {
		msg := "failed to encode HTTP request's JSON body: " + err.Error()
		return nil, temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err)
	}

	return bytes.NewReader(b), nil
}
