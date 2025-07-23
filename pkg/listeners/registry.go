package listeners

import (
	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/listeners/github"
	"github.com/tzrikka/timpani/pkg/listeners/slack"
)

// WebhookHandlers is a map of all the stateless webhook handlers that
// Timpani supports. The map keys correspond to Thrippy link template names.
var WebhookHandlers = map[string]listeners.WebhookHandlerFunc{
	"github-app-jwt":  github.WebhookHandler,
	"github-user-pat": github.WebhookHandler,
	"github-webhook":  github.WebhookHandler,
	"slack-bot-token": slack.WebhookHandler,
	"slack-oauth":     slack.WebhookHandler,
	"slack-oauth-gov": slack.WebhookHandler,
}

// ConnectionHandlers is a map of all the stateful connection handlers that
// Timpani supports. The map keys correspond to Thrippy link template names.
var ConnectionHandlers = map[string]listeners.ConnHandlerFunc{
	"slack-socket-mode": slack.ConnectionHandler,
}
