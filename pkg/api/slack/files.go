package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/pkg/metrics"
)

// https://docs.slack.dev/reference/methods/files.getUploadURLExternal/
//
// Note: according to Slack documentation, this should be an HTTP POST request, but
// that doesn't work in tests ("invalid_arguments" API error), so using GET instead.
func (a *API) FilesGetUploadURLExternalActivity(ctx context.Context, req slack.FilesGetUploadURLExternalRequest) (*slack.FilesGetUploadURLExternalResponse, error) {
	query := url.Values{}
	query.Set("length", strconv.Itoa(req.Length))
	query.Set("filename", req.Filename)
	if req.SnippetType != "" {
		query.Set("snippet_type", req.SnippetType)
	}
	if req.AltTxt != "" {
		query.Set("alt_txt", req.AltTxt)
	}

	t := time.Now().UTC()
	resp := new(slack.FilesGetUploadURLExternalResponse)
	if err := a.httpGet(ctx, slack.FilesGetUploadURLExternalActivityName, query, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.FilesGetUploadURLExternalActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.FilesGetUploadURLExternalActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.FilesGetUploadURLExternalActivityName, nil)
	return resp, nil
}

// TimpaniUploadExternalActivity uploads a file to Slack. It should be called after
// [FilesGetUploadURLExternalActivity] and before [FilesCompleteUploadExternalActivity].
func (a *API) TimpaniUploadExternalActivity(ctx context.Context, req slack.TimpaniUploadExternalRequest) error {
	t := time.Now().UTC()
	err := a.httpPostFile(ctx, req.URL, req.MimeType, req.Content)
	metrics.IncrementAPICallCounter(t, slack.TimpaniUploadExternalActivityName, err)
	return err
}

// https://docs.slack.dev/reference/methods/files.completeUploadExternal/
func (a *API) FilesCompleteUploadExternalActivity(ctx context.Context, req slack.FilesCompleteUploadExternalRequest) (*slack.FilesCompleteUploadExternalResponse, error) {
	t := time.Now().UTC()
	resp := new(slack.FilesCompleteUploadExternalResponse)
	if err := a.httpPost(ctx, slack.FilesCompleteUploadExternalActivityName, req, resp); err != nil {
		metrics.IncrementAPICallCounter(t, slack.FilesCompleteUploadExternalActivityName, err)
		return nil, err
	}

	if !resp.OK {
		metrics.IncrementAPICallCounter(t, slack.FilesCompleteUploadExternalActivityName, slackAPIError(resp, resp.Error))
		return nil, errors.New("Slack API error: " + resp.Error)
	}

	metrics.IncrementAPICallCounter(t, slack.FilesCompleteUploadExternalActivityName, nil)
	return resp, nil
}
