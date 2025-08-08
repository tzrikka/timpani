package slack

import (
	"context"
	"errors"
)

const (
	BookmarksAddName    = "slack.bookmarks.add"
	BookmarksEditName   = "slack.bookmarks.edit"
	BookmarksListName   = "slack.bookmarks.list"
	BookmarksRemoveName = "slack.bookmarks.remove"
)

// https://docs.slack.dev/reference/methods/bookmarks.add
type BookmarksAddRequest struct {
	ChannelID string `json:"channel_id"`
	Title     string `json:"title"`
	Type      string `json:"type"`

	Link        string `json:"link,omitempty"`
	Emoji       string `json:"emoji,omitempty"`
	EntityID    string `json:"entity_id,omitempty"`
	AccessLevel string `json:"access_level,omitempty"`
	ParentID    string `json:"parent_id,omitempty"`
}

// https://docs.slack.dev/reference/methods/bookmarks.add
type BookmarksAddResponse struct {
	slackResponse

	Bookmark map[string]any `json:"bookmark,omitempty"`
}

// https://docs.slack.dev/reference/methods/bookmarks.add
func (a *API) BookmarksAddActivity(ctx context.Context, req BookmarksAddRequest) (*BookmarksAddResponse, error) {
	resp := new(BookmarksAddResponse)
	if err := a.httpPost(ctx, BookmarksAddName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.edit
type BookmarksEditRequest struct {
	ChannelID  string `json:"channel_id"`
	BookmarkID string `json:"bookmark_id"`

	Title string `json:"title,omitempty"`
	Link  string `json:"link,omitempty"`
	Emoji string `json:"emoji,omitempty"`
}

// https://docs.slack.dev/reference/methods/bookmarks.edit
type BookmarksEditResponse struct {
	slackResponse

	Bookmark map[string]any `json:"bookmark,omitempty"`
}

// https://docs.slack.dev/reference/methods/bookmarks.edit
func (a *API) BookmarksEditActivity(ctx context.Context, req BookmarksEditRequest) (*BookmarksEditResponse, error) {
	resp := new(BookmarksEditResponse)
	if err := a.httpPost(ctx, BookmarksEditName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.list
type BookmarksListRequest struct {
	ChannelID string `json:"channel_id"`
}

// https://docs.slack.dev/reference/methods/bookmarks.list
type BookmarksListResponse struct {
	slackResponse

	Bookmarks []map[string]any `json:"bookmarks,omitempty"`
}

// https://docs.slack.dev/reference/methods/bookmarks.list
func (a *API) BookmarksListActivity(ctx context.Context, req BookmarksListRequest) (*BookmarksListResponse, error) {
	resp := new(BookmarksListResponse)
	if err := a.httpPost(ctx, BookmarksListName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}

// https://docs.slack.dev/reference/methods/bookmarks.remove
type BookmarksRemoveRequest struct {
	ChannelID  string `json:"channel_id"`
	BookmarkID string `json:"bookmark_id"`

	QuipSectionID string `json:"quip_section_id,omitempty"`
}

// https://docs.slack.dev/reference/methods/bookmarks.remove
type BookmarksRemoveResponse struct {
	slackResponse
}

// https://docs.slack.dev/reference/methods/bookmarks.remove
func (a *API) BookmarksRemoveActivity(ctx context.Context, req BookmarksRemoveRequest) (*BookmarksRemoveResponse, error) {
	resp := new(BookmarksRemoveResponse)
	if err := a.httpPost(ctx, BookmarksRemoveName, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New("Slack API error: " + resp.Error)
	}
	return resp, nil
}
