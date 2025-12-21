package slack

import (
	"context"
	"errors"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

// AuthTestActivity is based on:
// https://docs.slack.dev/reference/methods/auth.test/
func (a *API) AuthTestActivity(ctx context.Context) (*slack.AuthTestResponse, error) {
	resp := new(slack.AuthTestResponse)
	if err := a.httpPost(ctx, slack.AuthTestActivityName, nil, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
