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
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v3"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/tzrikka/timpani/internal/listeners"
	"github.com/tzrikka/timpani/pkg/api/slack"
)

// Run initializes the Temporal worker, and blocks.
func Run(l zerolog.Logger, cmd *cli.Command) error {
	addr := cmd.String("temporal-host-port")
	l.Info().Msgf("Temporal server address: %s", addr)

	c, err := client.Dial(client.Options{
		HostPort:  addr,
		Namespace: cmd.String("temporal-namespace"),
		Logger:    LogAdapter{zerolog: l},
	})
	if err != nil {
		return fmt.Errorf("failed to dial Temporal: %w", err)
	}
	defer c.Close()

	w := worker.New(c, cmd.String("temporal-task-queue"), worker.Options{})
	w.RegisterWorkflowWithOptions(waitForEventWorkflow, workflow.RegisterOptions{
		Name: listeners.WaitForEventWorkflow,
	})
	slack.Register(l, cmd, w)

	if err := w.Run(worker.InterruptCh()); err != nil {
		return fmt.Errorf("failed to start Temporal worker: %w", err)
	}

	return nil
}

// waitForEventWorkflow is a generic Temporal workflow that waits for a specific [Signal]
// call from an event listener. Timeouts are optional. This workflow supports cancellation.
func waitForEventWorkflow(ctx workflow.Context, req listeners.WaitForEventRequest) (map[string]any, error) {
	signal := fmt.Sprintf("%s.events.%s", strings.ToLower(req.Source), req.Name)

	// https://docs.temporal.io/develop/go/observability#visibility
	kw := temporal.NewSearchAttributeKeyKeyword("WaitingForSignal").ValueSet(signal)
	if err := workflow.UpsertTypedSearchAttributes(ctx, kw); err != nil {
		return nil, fmt.Errorf("failed to set workflow search attribute: %w", err)
	}

	childCtx, cancel := workflow.WithCancel(ctx)
	defer cancel()

	ch := workflow.GetSignalChannel(ctx, signal)
	payload := make(map[string]any)
	l := workflow.GetLogger(ctx)
	startTime := time.Now()

	selector := workflow.NewSelector(childCtx)
	selector.AddReceive(ch, func(c workflow.ReceiveChannel, _ bool) {
		c.Receive(ctx, &payload)
		l.Debug("received signal", "signal", signal, "duration", time.Since(startTime).String())
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
		l.Debug("waiting for signal without timeout", "signal", signal)
	} else {
		l.Debug("waiting for signal", "signal", signal, "timeout", req.Timeout)

		// Using a selector instead of ch.ReceiveWithTimeout() to support workflow cancellation.
		timer = workflow.NewTimer(ctx, timeout)
		selector.AddFuture(timer, func(_ workflow.Future) {
			l.Debug("timeout while waiting for signal", "signal", signal, "timeout", req.Timeout)
			err = fmt.Errorf("timeout (%s)", req.Timeout)
		})
	}

	selector.AddReceive(childCtx.Done(), func(workflow.ReceiveChannel, bool) {
		l.Error("workflow canceled while waiting for signal", "signal", signal, "error", childCtx.Err().Error())
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
		Logger:    LogAdapter{zerolog: *l},
	})
	if err != nil {
		return fmt.Errorf("client dial error: %w", err)
	}
	defer c.Close()

	// https://docs.temporal.io/list-filter
	// https://docs.temporal.io/search-attribute
	// https://docs.temporal.io/develop/go/observability#visibility
	query := "WorkflowType = '%s' AND WaitingForSignal = '%s' AND ExecutionStatus = '%s'"
	list, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
		Query: fmt.Sprintf(query, listeners.WaitForEventWorkflow, name, "Running"),
	})
	if err != nil {
		return fmt.Errorf("workflow search error: %w", err)
	}

	for _, info := range list.GetExecutions() {
		wid, rid := info.Execution.WorkflowId, info.Execution.RunId
		l.Debug().Str("signal", name).Str("workflow_id", wid).Str("run_id", rid).
			Msg("sending signal to Temporal workflow")
		if err := c.SignalWorkflow(ctx, wid, rid, name, payload); err != nil {
			return fmt.Errorf("signaling error: %w", err)
		}
	}

	return nil
}
