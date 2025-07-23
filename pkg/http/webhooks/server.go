package webhooks

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/credentials"

	intlis "github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/internal/thrippy"
	"github.com/tzrikka/timpani/pkg/listeners"
)

const (
	timeout = 3 * time.Second
	maxSize = 10 << 20 // 10 MiB.
)

type httpServer struct {
	httpPort     int             // To initialize the HTTP server.
	webhookLinks map[string]bool // Configured Thrippy link IDs.
	thrippyURL   *url.URL        // Optional passthrough for Thrippy OAuth.

	thrippyGRPCAddr string
	thrippyCreds    credentials.TransportCredentials

	temporal intlis.TemporalConfig // Destination for event notifications.
}

func NewHTTPServer(cmd *cli.Command) *httpServer {
	links := map[string]bool{}
	for _, fn := range cmd.FlagNames() {
		if strings.HasPrefix(fn, "thrippy-link-") {
			links[cmd.String(fn)] = true
		}
	}

	return &httpServer{
		httpPort:     cmd.Int("webhook-port"),
		webhookLinks: links,
		thrippyURL:   baseURL(cmd.String("thrippy-http-addr")),

		thrippyGRPCAddr: cmd.String("thrippy-server-addr"),
		thrippyCreds:    thrippy.SecureCreds(cmd),

		temporal: intlis.TemporalConfig{
			HostPort:  cmd.String("temporal-host-port"),
			Namespace: cmd.String("temporal-namespace"),
			TaskQueue: cmd.String("temporal-task-queue"),
		},
	}
}

// baseURL converts the given address (e.g. "localhost:14460") into a URL.
// If the address is empty, this function returns a nil reference.
func baseURL(addr string) *url.URL {
	if addr == "" {
		return nil
	}

	// Force an HTTP scheme.
	if strings.HasPrefix(addr, "https://") {
		addr = strings.Replace(addr, "https://", "http://", 1)
	}
	if !strings.HasPrefix(addr, "http://") {
		addr = "http://" + addr
	}

	// Strip any suffix after the address.
	u, err := url.Parse(addr)
	if err != nil {
		return nil
	}
	if u.Host == "" {
		return nil
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u
}

// Run starts an HTTP server to expose webhooks, and blocks forever.
func (s *httpServer) Run() {
	http.HandleFunc("GET /webhook/{id...}", s.webhookHandler)
	http.HandleFunc("POST /webhook/{id...}", s.webhookHandler)

	if s.thrippyURL != nil {
		log.Info().Msgf("HTTP passthrough for Thrippy OAuth callbacks: %s", s.thrippyURL)
		http.HandleFunc("GET /callback", s.thrippyHandler)
		http.HandleFunc("GET /start", s.thrippyHandler)
		http.HandleFunc("POST /start", s.thrippyHandler)
		http.HandleFunc("GET /success", s.thrippyHandler)
	}

	server := &http.Server{
		Addr:         net.JoinHostPort("", strconv.Itoa(s.httpPort)),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	log.Info().Msgf("HTTP server listening on port %d", s.httpPort)
	_ = server.ListenAndServe()
}

// webhookHandler checks and processes incoming asynchronous
// event notifications over HTTP from third-party services.
func (s *httpServer) webhookHandler(w http.ResponseWriter, r *http.Request) {
	l := log.With().Str("http_method", r.Method).Str("url_path", r.URL.EscapedPath()).Logger()
	if r.Method == http.MethodPost {
		l = l.With().Str("content_type", r.Header.Get("Content-Type")).Logger()
	}
	l.Info().Msg("received HTTP request")

	linkID, pathSuffix, statusCode := parseURL(l, r)
	if statusCode != http.StatusOK {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	l = l.With().Str("link_id", linkID).Logger()
	if pathSuffix != "" {
		l = l.With().Str("path_suffix", pathSuffix).Logger()
	}

	if _, ok := s.webhookLinks[linkID]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	template, secrets, err := s.linkData(r.Context(), linkID)
	if statusCode := checkLinkData(l, template, secrets, err); statusCode != http.StatusOK {
		w.WriteHeader(statusCode)
		return
	}

	raw, decoded, err := parseBody(w, r)
	if err != nil {
		l.Warn().Err(err).Msg("bad request: JSON decoding error")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(raw))
	_ = r.ParseForm()

	// Forward the request's data to a service-specific handler.
	l = l.With().Str("template", template).Logger()
	f, ok := listeners.WebhookHandlers[template]
	if !ok {
		l.Warn().Msg("bad request: unsupported link template for webhooks")
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	statusCode = f(l.WithContext(r.Context()), w, intlis.RequestData{
		PathSuffix:  pathSuffix,
		Headers:     r.Header,
		WebForm:     r.Form,
		RawPayload:  raw,
		JSONPayload: decoded,
		LinkSecrets: secrets,
		Temporal:    s.temporal,
	})
	if statusCode != 0 {
		w.WriteHeader(statusCode)
	}
}

// parseURL extracts the Thrippy link ID from the request's URL path.
// The path may contain an opaque suffix after the ID, separated by a slash,
// for third-party services that support/require multiple URLs per connection.
func parseURL(l zerolog.Logger, r *http.Request) (string, string, int) {
	id := r.PathValue("id")
	if id == "" {
		l.Warn().Msg("bad request: missing ID")
		return "", "", http.StatusBadRequest
	}

	suffix := ""
	if strings.Contains(id, "/") {
		parts := strings.SplitN(id, "/", 2)
		id = parts[0]
		suffix = parts[1]
	}

	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn().Err(err).Msg("bad request: ID is an invalid short UUID")
		return "", "", http.StatusNotFound
	}

	return id, suffix, http.StatusOK
}

// parseBody tries to parse the given HTTP request body as JSON.
// It also returns the raw payload to support authenticity checks.
// If the request is not a POST with a JSON content type, it returns nil.
func parseBody(w http.ResponseWriter, r *http.Request) ([]byte, map[string]any, error) {
	if r.Method != http.MethodPost {
		return nil, nil, nil
	}

	raw, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSize))
	if err != nil {
		return nil, nil, err
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return raw, nil, nil
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return nil, nil, err
	}

	return raw, decoded, nil
}
