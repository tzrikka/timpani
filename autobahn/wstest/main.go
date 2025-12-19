// Wstest tests Timpani's [WebSocket client] against
// the fuzzing server of the [Autobahn Testsuite].
//
// [WebSocket client]: https://pkg.go.dev/github.com/tzrikka/timpani/pkg/websocket
// [Autobahn Testsuite]: https://github.com/crossbario/autobahn-testsuite
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/tzrikka/timpani/internal/logger"
	"github.com/tzrikka/timpani/pkg/websocket"
)

const (
	baseURL = "ws://127.0.0.1:9001"
	agent   = "timpani"
)

func main() {
	n := getCaseCount()
	slog.Info("case count", slog.Int("n", n+1))

	// Not implemented in Timpani (so excluded in "config/fuzzingserver.json"):
	//   - 6.4.*: Fail-fast on invalid UTF-8 frames,
	//   - 12.* and 13.*: WebSocket compression.
	for i := range n {
		runCase(i + 1)
	}

	updateReports()
}

func dial(url string) (*websocket.Conn, error) {
	return websocket.Dial(context.Background(), url)
}

// getCaseCount retrieves the number of enabled test cases from
// the Autobahn fuzzing server, using a WebSocket request.
func getCaseCount() int {
	conn, err := dial(baseURL + "/getCaseCount")
	if err != nil {
		logger.FatalError("dial error", err)
	}

	msg, ok := <-conn.IncomingMessages()
	if !ok {
		slog.Debug("connection closed")
		return 0
	}

	n, err := strconv.Atoi(string(msg.Data))
	if err != nil {
		logger.FatalError("invalid test case count", err)
	}

	return n
}

// updateReports instructs the Autobahn fuzzing server to generate/update
// all the HTML and JSON files for all the test-case results.
func updateReports() {
	slog.Info("updating reports")

	url := fmt.Sprintf("%s/updateReports?agent=%s", baseURL, agent)
	if _, err := dial(url); err != nil {
		logger.FatalError("dial error", err)
	}
}

func runCase(i int) {
	l := slog.With(slog.Int("case", i))
	l.Info("starting test")

	conn, err := dial(fmt.Sprintf("%s/runCase?case=%d&agent=%s", baseURL, i, agent))
	if err != nil {
		logger.FatalError("dial error", err)
	}

	// Echo loop.
	for {
		msg := <-conn.IncomingMessages()
		if msg.Data == nil {
			l.Debug("connection closed")
			break
		}

		l = l.With(slog.String("opcode", msg.Opcode.String()))
		l.Info("received message", slog.Int("length", len(msg.Data)))

		switch msg.Opcode {
		case websocket.OpcodeText:
			err = <-conn.SendTextMessage(msg.Data)
		case websocket.OpcodeBinary:
			err = <-conn.SendBinaryMessage(msg.Data)
		default:
			l.Error("unexpected opcode in data message")
			os.Exit(1)
		}

		if err != nil {
			l.Error("echo error", slog.Any("error", err))
			conn.Close(websocket.StatusNormalClosure)
		}
	}
}
