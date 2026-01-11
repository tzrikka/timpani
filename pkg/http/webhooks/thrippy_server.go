package webhooks

import (
	"context"
	"io"
	"log/slog"
	"net/http"
)

// thrippyHandler passes-through incoming HTTP requests (OAuth callbacks),
// as a proxy, to a local Thrippy server. This allows Timpani and Thrippy to
// share a single HTTP tunnel when running together in a local development setup.
func (s *HTTPServer) thrippyHandler(w http.ResponseWriter, r *http.Request) {
	l := slog.With(slog.String("http_method", r.Method), slog.String("url_path", r.URL.EscapedPath()))
	l.Info("passing-through HTTP request to Thrippy")

	// Adjust the original URL to the Thrippy server's base URL.
	r.URL.Scheme = s.thrippyURL.Scheme
	r.URL.Host = s.thrippyURL.Host

	// Construct the proxy request.
	ctx, cancel := context.WithTimeout(r.Context(), Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, r.Method, r.URL.String(), r.Body)
	if err != nil {
		l.Error("failed to construct Thrippy proxy request", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req.Header = r.Header.Clone()

	// Send the proxy request.
	client := &http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse // Let the client handle all redirects.
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		l.Error("failed to send Thrippy proxy request", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Relay Thrippy's response back to the client.
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		l.Error("failed to copy Thrippy response body", slog.Any("error", err))
	}
}
