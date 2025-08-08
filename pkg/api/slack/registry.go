package slack

import (
	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

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

	registerActivity(w, a.BookmarksAddActivity, BookmarksAddName)
	registerActivity(w, a.BookmarksEditActivity, BookmarksEditName)
	registerActivity(w, a.BookmarksListActivity, BookmarksListName)
	registerActivity(w, a.BookmarksRemoveActivity, BookmarksRemoveName)

	registerActivity(w, a.ChatDeleteActivity, ChatDeleteName)
	registerActivity(w, a.ChatGetPermalinkActivity, ChatGetPermalinkName)
	registerActivity(w, a.ChatPostEphemeralActivity, ChatPostEphemeralName)
	registerActivity(w, a.ChatPostMessageActivity, ChatPostMessageName)
	registerActivity(w, a.ChatUpdateActivity, ChatUpdateName)

	registerActivity(w, a.ConversationsArchiveActivity, ConversationsArchiveName)
	registerActivity(w, a.ConversationsCloseActivity, ConversationsCloseName)
	registerActivity(w, a.ConversationsCreateActivity, ConversationsCreateName)
	registerActivity(w, a.ConversationsHistoryActivity, ConversationsHistoryName)
	registerActivity(w, a.ConversationsInfoActivity, ConversationsInfoName)
	registerActivity(w, a.ConversationsInviteActivity, ConversationsInviteName)
	registerActivity(w, a.ConversationsJoinActivity, ConversationsJoinName)
	registerActivity(w, a.ConversationsKickActivity, ConversationsKickName)
	registerActivity(w, a.ConversationsLeaveActivity, ConversationsLeaveName)
	registerActivity(w, a.ConversationsListActivity, ConversationsListName)
	registerActivity(w, a.ConversationsMembersActivity, ConversationsMembersName)
	registerActivity(w, a.ConversationsOpenActivity, ConversationsOpenName)
	registerActivity(w, a.ConversationsRenameActivity, ConversationsRenameName)
	registerActivity(w, a.ConversationsRepliesActivity, ConversationsRepliesName)
	registerActivity(w, a.ConversationsSetPurposeActivity, ConversationsSetPurposeName)
	registerActivity(w, a.ConversationsSetTopicActivity, ConversationsSetTopicName)
	registerActivity(w, a.ConversationsUnarchiveActivity, ConversationsUnarchiveName)

	registerActivity(w, a.ReactionsAddActivity, ReactionsAddName)
	registerActivity(w, a.ReactionsGetActivity, ReactionsGetName)
	registerActivity(w, a.ReactionsListActivity, ReactionsListName)
	registerActivity(w, a.ReactionsRemoveActivity, ReactionsRemoveName)

	registerActivity(w, a.UsersConversationsActivity, UsersConversationsName)
	registerActivity(w, a.UsersGetPresenceActivity, UsersGetPresenceName)
	registerActivity(w, a.UsersIdentityActivity, UsersIdentityName)
	registerActivity(w, a.UsersInfoActivity, UsersInfoName)
	registerActivity(w, a.UsersListActivity, UsersListName)
	registerActivity(w, a.UsersLookupByEmailActivity, UsersLookupByEmailName)
	registerActivity(w, a.UsersProfileGetActivity, UsersProfileGetName)

	registerWorkflow(w, a.TimpaniPostApprovalWorkflow, TimpaniPostApprovalName)
}

func registerActivity(w worker.Worker, f any, name string) {
	w.RegisterActivityWithOptions(f, activity.RegisterOptions{Name: name})
}

func registerWorkflow(w worker.Worker, f any, name string) {
	w.RegisterWorkflowWithOptions(f, workflow.RegisterOptions{Name: name})
}
