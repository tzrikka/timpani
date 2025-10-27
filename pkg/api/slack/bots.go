package slack

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://docs.slack.dev/reference/methods/bots.info/
func (a *API) BotsInfoActivity(ctx context.Context, req slack.BotsInfoRequest) (*slack.BotsInfoResponse, error) {
	query := url.Values{}
	query.Set("bot", req.Bot)
	if req.TeamID != "" {
		query.Set("team_id", req.TeamID)
	}

	t := time.Now().UTC()
	resp := new(slack.BotsInfoResponse)
	if err := a.httpGet(ctx, slack.BotsInfoActivityName, query, resp); err != nil {
		metrics.CountAPICall(t, slack.BotsInfoActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.CountAPICall(t, slack.BotsInfoActivityName, errors.New(resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.CountAPICall(t, slack.BotsInfoActivityName, nil)
	return resp, nil
}
