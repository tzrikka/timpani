// Package client provides a simple, generic HTTP client
// for sending various API requests to external services,
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
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/temporal"
)

const (
	Timeout = 3 * time.Second
	MaxSize = 3 << 20 // 3 MiB.

	AcceptJSON = "application/json"
	AcceptText = "text/plain"

	ContentForm = "application/x-www-form-urlencoded"
	ContentJSON = "application/json; charset=utf-8"
)

// HTTPRequest sends an HTTP request to an external API service.
//
// The queryOrBody parameter may be nil, [url.Values], a []byte slice,
// or any struct that can be encoded as JSON.
//
// Some errors (failure to construct a request or decode a response body)
// are returned as non-retryable [temporal.ApplicationError]s.
//
// On HTTP 429 (Too Many Requests) responses, the third return value
// contains the number of seconds to wait before retrying the request.
//
// [temporal.ApplicationError]: https://pkg.go.dev/go.temporal.io/temporal#ApplicationError
func HTTPRequest(ctx context.Context, method, apiURL, auth, accept, contentType string, queryOrBody any) ([]byte, http.Header, int, error) {
	// Construct the request.
	if method == http.MethodGet || method == http.MethodDelete {
		if query, ok := queryOrBody.(url.Values); ok && len(query) > 0 {
			apiURL = fmt.Sprintf("%s?%s", apiURL, query.Encode())
		}
	}

	reqBody, err := requestBody(method, queryOrBody)
	if err != nil {
		return nil, nil, 0, err
	}

	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, method, apiURL, reqBody)
	if err != nil {
		msg := "failed to construct HTTP request: " + err.Error()
		return nil, nil, 0, temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err)
	}

	// Set HTTP headers for auth and request/response MIME types.
	if pair, found := strings.CutPrefix(auth, "Basic "); found {
		if user, pass, found := strings.Cut(pair, ":"); found {
			req.SetBasicAuth(user, pass)
		}
	} else if auth != "" {
		req.Header.Set("Authorization", "Bearer "+auth)
	}

	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if method != http.MethodGet && method != http.MethodDelete {
		req.Header.Set("Content-Type", contentType)
	}

	// Send the request, and read the response.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, MaxSize))
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to read HTTP response body: %w", err)
	}

	return parseResponse(resp, respBody)
}

func requestBody(method string, queryOrBody any) (io.Reader, error) {
	if method == http.MethodGet || method == http.MethodDelete {
		return http.NoBody, nil
	}

	if rawBytes, ok := queryOrBody.([]byte); ok {
		return bytes.NewReader(rawBytes), nil
	}

	// HTTP PATCH, POST, or PUT with a JSON body.
	jsonBody, err := json.Marshal(queryOrBody)
	if err != nil {
		msg := "failed to encode HTTP request's JSON body: " + err.Error()
		return nil, temporal.NewNonRetryableApplicationError(msg, fmt.Sprintf("%T", err), err)
	}

	return bytes.NewReader(jsonBody), nil
}

func parseResponse(resp *http.Response, body []byte) ([]byte, http.Header, int, error) {
	if resp.StatusCode < http.StatusBadRequest {
		return body, resp.Header, 0, nil
	}

	var retryAfter float64
	msg := resp.Status

	// Rate-limit handling, based on: https://datatracker.ietf.org/doc/html/rfc6585#section-4.1
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter, _ = strconv.ParseFloat(resp.Header.Get("Retry-After"), 64)
	}

	// Additional GitHub-specific rate-limit handling, based on:
	// https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api?apiVersion=2022-11-28
	if retryAfter == 0 && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests) {
		if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining == "0" {
			if reset, err := strconv.ParseInt(resp.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
				retryAfter = math.Ceil(time.Until(time.Unix(reset, 0)).Seconds())
			}
		}
	}

	secs := int(math.Max(0, retryAfter))
	if secs > 0 {
		msg += fmt.Sprintf(" (retry after %d seconds)", secs)
	}

	if len(body) > 0 {
		msg += ": " + string(body)
	}

	return nil, nil, secs, errors.New(msg)
}
