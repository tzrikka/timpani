package slack

import (
	"context"
	"errors"
	"net/url"
)

const (
	BotsInfoName = "slack.bots.info"
)

// https://docs.slack.dev/reference/methods/bots.info
type BotsInfoRequest struct {
	Bot string `json:"bot"`

	TeamID string `json:"team_id,omitempty"`
}

// https://docs.slack.dev/reference/methods/bots.info
type BotsInfoResponse struct {
	slackResponse

	Bot map[string]any `json:"bot,omitempty"`
}

// https://docs.slack.dev/reference/methods/bots.info
func (a *API) BotsInfoActivity(ctx context.Context, req BotsInfoRequest) (*BotsInfoResponse, error) {
	query := url.Values{}
	query.Set("bot", req.Bot)
	query.Set("team_id", req.TeamID)

	resp := new(BotsInfoResponse)
	if err := a.httpGet(ctx, BotsInfoName, query, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
