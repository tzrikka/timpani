// Package temporal initializes a Temporal worker that provides activities
// and workflows wrapping the APIs of various third-party services which are
// defined and implemented under the [pkg/api] and [pkg/listeners] packages.
//
// [pkg/api]: https://pkg.go.dev/github.com/tzrikka/timpani/pkg/api
// [pkg/listeners]: https://pkg.go.dev/github.com/tzrikka/timpani/pkg/listeners
package temporal

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/api/bitbucket"
	"github.com/tzrikka/timpani/pkg/api/github"
	"github.com/tzrikka/timpani/pkg/api/slack"
)

// Run initializes the Temporal worker, and blocks to keep it running.
func Run(l zerolog.Logger, cmd *cli.Command) error {
	addr := cmd.String("temporal-address")
	l.Info().Msgf("Temporal server address: %s", addr)

	c, err := client.Dial(client.Options{
		HostPort:  addr,
		Namespace: cmd.String("temporal-namespace"),
		Logger:    LogAdapter{zerolog: l.With().CallerWithSkipFrameCount(4).Logger()},
	})
	if err != nil {
		return fmt.Errorf("failed to dial Temporal: %w", err)
	}
	defer c.Close()

	w := worker.New(c, cmd.String("temporal-task-queue"), worker.Options{})
	w.RegisterWorkflowWithOptions(waitForEventWorkflow, workflow.RegisterOptions{
		Name: listeners.WaitForEventWorkflow,
	})
	bitbucket.Register(l, cmd, w)
	github.Register(l, cmd, w)
	slack.Register(l, cmd, w)

	if err := w.Run(worker.InterruptCh()); err != nil {
		return fmt.Errorf("failed to start Temporal worker: %w", err)
	}

	return nil
}

// waitForEventWorkflow is a generic Temporal workflow that waits for a specific [Signal]
// call from an event listener. Timeouts are optional. This workflow supports cancellation.
func waitForEventWorkflow(ctx workflow.Context, req listeners.WaitForEventRequest) (map[string]any, error) {
	// https://docs.temporal.io/develop/go/observability#visibility
	sa := temporal.NewSearchAttributeKeyKeywordList("WaitingForSignals").ValueSet([]string{req.Signal})
	if err := workflow.UpsertTypedSearchAttributes(ctx, sa); err != nil {
		return nil, fmt.Errorf("failed to set workflow search attribute: %w", err)
	}

	childCtx, cancel := workflow.WithCancel(ctx)
	defer cancel()

	ch := workflow.GetSignalChannel(ctx, req.Signal)
	payload := make(map[string]any)
	l := workflow.GetLogger(ctx)
	startTime := time.Now()

	selector := workflow.NewSelector(childCtx)
	selector.AddReceive(ch, func(c workflow.ReceiveChannel, _ bool) {
		c.Receive(ctx, &payload)
		l.Debug("received signal", "signal", req.Signal, "duration", time.Since(startTime).String())
	})

	if req.Timeout == "" {
		req.Timeout = "0s"
	}
	timeout, err := time.ParseDuration(req.Timeout)
	if err != nil {
		return nil, err
	}

	var timer workflow.Future
	if timeout == 0 {
		l.Debug("waiting for signal without timeout", "signal", req.Signal)
	} else {
		l.Debug("waiting for signal", "signal", req.Signal, "timeout", req.Timeout)

		// Using a selector instead of ch.ReceiveWithTimeout() to support workflow cancellation.
		timer = workflow.NewTimer(ctx, timeout)
		selector.AddFuture(timer, func(_ workflow.Future) {
			l.Debug("timeout while waiting for signal", "signal", req.Signal, "timeout", req.Timeout)
			err = fmt.Errorf("timeout (%s)", req.Timeout)
		})
	}

	selector.AddReceive(childCtx.Done(), func(workflow.ReceiveChannel, bool) {
		l.Error("workflow canceled while waiting for signal", "error", childCtx.Err(), "signal", req.Signal)
	})

	selector.Select(ctx)

	switch {
	case childCtx.Err() != nil:
		return nil, childCtx.Err()
	case err != nil:
		return nil, err
	default:
		return payload, nil
	}
}

// Signal sends a specific payload, which was received as an asynchronous event
// notification, to all (zero of more) Temporal workflows that are waiting for it.
//
// The ctx parameter is expected to have a ZeroLog logger attached to it:
//
//	ctx = l.WithContext(ctx)
func Signal(ctx context.Context, cfg listeners.TemporalConfig, name string, payload map[string]any) error {
	l := zerolog.Ctx(ctx)

	c, err := client.Dial(client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
		Logger:    LogAdapter{zerolog: l.With().CallerWithSkipFrameCount(4).Logger()},
	})
	if err != nil {
		return fmt.Errorf("client dial error: %w", err)
	}
	defer c.Close()

	// https://docs.temporal.io/list-filter
	// https://docs.temporal.io/search-attribute
	// https://docs.temporal.io/develop/go/observability#visibility
	name = sanitizeSignalName(l, name)
	list, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: fmt.Sprintf("WaitingForSignals IN ('%s') AND ExecutionStatus = '%s'", name, "Running"),
	})
	if err != nil {
		return fmt.Errorf("workflow search error: %w", err)
	}

	for _, info := range list.GetExecutions() {
		wid, rid := info.Execution.WorkflowId, info.Execution.RunId
		l.Info().Str("signal", name).Str("workflow_id", wid).Str("run_id", rid).
			Msg("sending signal to Temporal workflow")
		if err := c.SignalWorkflow(ctx, wid, rid, name, payload); err != nil {
			return fmt.Errorf("signaling error: %w", err)
		}
	}

	return nil
}

var ForbiddenSignalNameChars = regexp.MustCompile("[^0-9A-Za-z_.]")

// sanitizeSignalName ensures that signal names (generated from incoming events)
// cannot manipulate Timpani's Temporal query in the [Signal] function.
func sanitizeSignalName(l *zerolog.Logger, name string) string {
	safeName := ForbiddenSignalNameChars.ReplaceAllString(name, "_")
	if len(safeName) > 100 {
		safeName = safeName[:100]
	}

	if name != safeName {
		l.Warn().Str("original", name).Str("sanitized", safeName).
			Msg("signal name contained forbidden characters")
	}

	return safeName
}
