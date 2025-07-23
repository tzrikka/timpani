package temporal

import (
	"fmt"

	"github.com/rs/zerolog"
)

// LogAdapter implements Temporal's [Logger] interface.
//
// [Logger]: https://pkg.go.dev/go.temporal.io/sdk/log#Logger
type LogAdapter struct {
	zerolog zerolog.Logger
}

func (a LogAdapter) Debug(msg string, keyvals ...any) {
	logMessageWithAttributes(a.zerolog.Debug(), msg, keyvals)
}

func (a LogAdapter) Info(msg string, keyvals ...any) {
	logMessageWithAttributes(a.zerolog.Info(), msg, keyvals)
}

func (a LogAdapter) Warn(msg string, keyvals ...any) {
	logMessageWithAttributes(a.zerolog.Warn(), msg, keyvals)
}

func (a LogAdapter) Error(msg string, keyvals ...any) {
	logMessageWithAttributes(a.zerolog.Error().Stack(), msg, keyvals)
}

func logMessageWithAttributes(e *zerolog.Event, msg string, keyvals ...any) {
	for i, kv := range keyvals {
		as, ok := kv.([]any)
		if !ok {
			e = e.Any(fmt.Sprintf("attr_%d", i), kv)
			continue
		}
		for len(as) > 1 {
			e = e.Any(as[0].(string), as[1])
			as = as[2:]
		}
		if len(as) > 0 {
			e = e.Str(as[0].(string), "")
		}
	}
	e.Msg(msg)
}
