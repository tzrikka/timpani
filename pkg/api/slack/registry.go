package slack

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/tzrikka/timpani-api/pkg/slack"
	"github.com/tzrikka/timpani/internal/thrippy"
)

type API struct {
	thrippy thrippy.LinkClient
}

// Register exposes Temporal activities and workflows via the Timpani worker.
func Register(l zerolog.Logger, cmd *cli.Command, w worker.Worker) {
	id, ok := thrippy.LinkID(l, cmd, "Slack")
	if !ok {
		return
	}

	a := API{thrippy: thrippy.NewLinkClient(id, cmd)}

	registerActivity(w, a.AuthTestActivity, slack.AuthTestActivityName)

	registerActivity(w, a.BookmarksAddActivity, slack.BookmarksAddActivityName)
	registerActivity(w, a.BookmarksEditActivity, slack.BookmarksEditActivityName)
	registerActivity(w, a.BookmarksListActivity, slack.BookmarksListActivityName)
	registerActivity(w, a.BookmarksRemoveActivity, slack.BookmarksRemoveActivityName)

	registerActivity(w, a.BotsInfoActivity, slack.BotsInfoActivityName)

	registerActivity(w, a.ChatDeleteActivity, slack.ChatDeleteActivityName)
	registerActivity(w, a.ChatGetPermalinkActivity, slack.ChatGetPermalinkActivityName)
	registerActivity(w, a.ChatPostEphemeralActivity, slack.ChatPostEphemeralActivityName)
	registerActivity(w, a.ChatPostMessageActivity, slack.ChatPostMessageActivityName)
	registerActivity(w, a.ChatUpdateActivity, slack.ChatUpdateActivityName)

	registerActivity(w, a.ConversationsArchiveActivity, slack.ConversationsArchiveActivityName)
	registerActivity(w, a.ConversationsCloseActivity, slack.ConversationsCloseActivityName)
	registerActivity(w, a.ConversationsCreateActivity, slack.ConversationsCreateActivityName)
	registerActivity(w, a.ConversationsHistoryActivity, slack.ConversationsHistoryActivityName)
	registerActivity(w, a.ConversationsInfoActivity, slack.ConversationsInfoActivityName)
	registerActivity(w, a.ConversationsInviteActivity, slack.ConversationsInviteActivityName)
	registerActivity(w, a.ConversationsJoinActivity, slack.ConversationsJoinActivityName)
	registerActivity(w, a.ConversationsKickActivity, slack.ConversationsKickActivityName)
	registerActivity(w, a.ConversationsLeaveActivity, slack.ConversationsLeaveActivityName)
	registerActivity(w, a.ConversationsListActivity, slack.ConversationsListActivityName)
	registerActivity(w, a.ConversationsMembersActivity, slack.ConversationsMembersActivityName)
	registerActivity(w, a.ConversationsOpenActivity, slack.ConversationsOpenActivityName)
	registerActivity(w, a.ConversationsRenameActivity, slack.ConversationsRenameActivityName)
	registerActivity(w, a.ConversationsRepliesActivity, slack.ConversationsRepliesActivityName)
	registerActivity(w, a.ConversationsSetPurposeActivity, slack.ConversationsSetPurposeActivityName)
	registerActivity(w, a.ConversationsSetTopicActivity, slack.ConversationsSetTopicActivityName)

	registerActivity(w, a.FilesCompleteUploadExternalActivity, slack.FilesCompleteUploadExternalActivityName)
	registerActivity(w, a.FilesDeleteActivity, slack.FilesDeleteActivityName)
	registerActivity(w, a.FilesGetUploadURLExternalActivity, slack.FilesGetUploadURLExternalActivityName)
	registerActivity(w, a.TimpaniUploadExternalActivity, slack.TimpaniUploadExternalActivityName)

	registerActivity(w, a.ReactionsAddActivity, slack.ReactionsAddActivityName)
	registerActivity(w, a.ReactionsGetActivity, slack.ReactionsGetActivityName)
	registerActivity(w, a.ReactionsListActivity, slack.ReactionsListActivityName)
	registerActivity(w, a.ReactionsRemoveActivity, slack.ReactionsRemoveActivityName)

	registerActivity(w, a.UsersConversationsActivity, slack.UsersConversationsActivityName)
	registerActivity(w, a.UsersGetPresenceActivity, slack.UsersGetPresenceActivityName)
	registerActivity(w, a.UsersInfoActivity, slack.UsersInfoActivityName)
	registerActivity(w, a.UsersListActivity, slack.UsersListActivityName)
	registerActivity(w, a.UsersLookupByEmailActivity, slack.UsersLookupByEmailActivityName)
	registerActivity(w, a.UsersProfileGetActivity, slack.UsersProfileGetActivityName)

	registerWorkflow(w, a.TimpaniPostApprovalWorkflow, slack.TimpaniPostApprovalWorkflowName)
}

func registerActivity(w worker.Worker, f any, name string) {
	w.RegisterActivityWithOptions(f, activity.RegisterOptions{Name: name})
}

func registerWorkflow(w worker.Worker, f any, name string) {
	w.RegisterWorkflowWithOptions(f, workflow.RegisterOptions{Name: name})
}
