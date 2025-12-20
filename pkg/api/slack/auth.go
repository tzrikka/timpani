package slack

import (
	"context"
	"errors"
	"time"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// AuthTestActivity is based on:
// https://docs.slack.dev/reference/methods/auth.test/
func (a *API) AuthTestActivity(ctx context.Context) (*slack.AuthTestResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.AuthTestResponse)
	if err := a.httpPost(ctx, slack.AuthTestActivityName, nil, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.AuthTestActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.AuthTestActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.AuthTestActivityName, nil)
	return resp, nil
}
