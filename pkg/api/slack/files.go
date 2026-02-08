package slack

import (
	"context"
	"errors"
	"net/url"
	"strconv"

	"go.temporal.io/sdk/temporal"

	"github.com/tzrikka/timpani-api/pkg/slack"
)

// FilesGetUploadURLExternalActivity is based on:
// https://docs.slack.dev/reference/methods/files.getUploadURLExternal/
//
// Note: according to Slack documentation, this should be an HTTP POST request, but
// that doesn't work in tests ("invalid_arguments" API error), so using GET instead.
func (a *API) FilesGetUploadURLExternalActivity(
	ctx context.Context,
	req slack.FilesGetUploadURLExternalRequest,
) (*slack.FilesGetUploadURLExternalResponse, error) {
	query := url.Values{}
	query.Set("length", strconv.Itoa(req.Length))
	query.Set("filename", req.Filename)
	if req.SnippetType != "" {
		query.Set("snippet_type", req.SnippetType)
	}
	if req.AltTxt != "" {
		query.Set("alt_txt", req.AltTxt)
	}

	resp := new(slack.FilesGetUploadURLExternalResponse)
	if err := a.httpGet(ctx, slack.FilesGetUploadURLExternalActivityName, query, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// TimpaniUploadExternalActivity uploads a file to Slack. It should be called after
// [FilesGetUploadURLExternalActivity] and before [FilesCompleteUploadExternalActivity].
func (a *API) TimpaniUploadExternalActivity(ctx context.Context, req slack.TimpaniUploadExternalRequest) error {
	return a.httpPostFile(ctx, req.URL, req.MimeType, req.Content)
}

// FilesCompleteUploadExternalActivity is based on:
// https://docs.slack.dev/reference/methods/files.completeUploadExternal/
func (a *API) FilesCompleteUploadExternalActivity(
	ctx context.Context,
	req slack.FilesCompleteUploadExternalRequest,
) (*slack.FilesCompleteUploadExternalResponse, error) {
	resp := new(slack.FilesCompleteUploadExternalResponse)
	if err := a.httpPost(ctx, slack.FilesCompleteUploadExternalActivityName, req, resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// FilesDeleteActivity is based on:
// https://docs.slack.dev/reference/methods/files.delete/
func (a *API) FilesDeleteActivity(ctx context.Context, req slack.FilesDeleteRequest) (*slack.FilesDeleteResponse, error) {
	resp := new(slack.FilesDeleteResponse)
	if err := a.httpPost(ctx, slack.FilesDeleteActivityName, req, resp); err != nil {
		return nil, err
	}

	if resp.Error == "file_not_found" {
		return nil, temporal.NewNonRetryableApplicationError(resp.Error, "SlackAPIError", nil, req.File)
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
