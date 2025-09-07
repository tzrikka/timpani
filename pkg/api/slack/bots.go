package slack

import (
	"context"
	"errors"
	"net/url"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

// https://docs.slack.dev/reference/methods/bots.info/
func (a *API) BotsInfoActivity(ctx context.Context, req slack.BotsInfoRequest) (*slack.BotsInfoResponse, error) {
	query := url.Values{}
	query.Set("bot", req.Bot)
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}

	resp := new(slack.BotsInfoResponse)
	if err := a.httpGet(ctx, slack.BotsInfoActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	return resp, nil
}
