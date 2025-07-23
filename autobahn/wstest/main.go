// wstest tests Timpani's [WebSocket client] against
// the fuzzing server of the [Autobahn Testsuite].
//
// [WebSocket client]: https://pkg.go.dev/github.com/tzrikka/timpani/pkg/websocket
// [Autobahn Testsuite]: https://github.com/crossbario/autobahn-testsuite
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/tzrikka/timpani/pkg/websocket"
)

const (
	baseURL = "ws://127.0.0.1:9001"
	agent   = "timpani"
)

func main() {
	initZeroLog()
	n := getCaseCount()
	log.Logger.Info().Int("n", n+1).Msg("case count")

	// Not implemented in Timpani (so excluded in "config/fuzzingserver.json"):
	// - 6.4.*: Fail-fast on invalid UTF-8 frames
	// - 12.* and 13.*: WebSocket compression
	for i := range n {
		runCase(i + 1)
	}

	updateReports()
}

func initZeroLog() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05.000",
	}).With().Caller().Logger()
}

func dial(url string) (*websocket.Conn, error) {
	return websocket.Dial(log.Logger.WithContext(context.Background()), url)
}

// getCaseCount retrieves the number of enabled test cases from
// the Autobahn fuzzing server, using a WebSocket request.
func getCaseCount() int {
	conn, err := dial(baseURL + "/getCaseCount")
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("dial error")
	}

	msg, ok := <-conn.IncomingMessages()
	if !ok {
		log.Logger.Debug().Msg("connection closed")
		return 0
	}

	n, err := strconv.Atoi(string(msg.Data))
	if err != nil {
		log.Logger.Fatal().Err(err).Msg("invalid test case count")
	}

	return n
}

// updateReports instructs the Autobahn fuzzing server to generate/update
// all the HTML and JSON files for all the test-case results.
func updateReports() {
	log.Logger.Info().Msg("updating reports")

	url := fmt.Sprintf("%s/updateReports?agent=%s", baseURL, agent)
	if _, err := dial(url); err != nil {
		log.Logger.Fatal().Err(err).Msg("dial error")
	}
}

func runCase(i int) {
	l := log.Logger.With().Int("case", i).Logger()
	l.Info().Msg("starting test")

	conn, err := dial(fmt.Sprintf("%s/runCase?case=%d&agent=%s", baseURL, i, agent))
	if err != nil {
		l.Fatal().Err(err).Msg("dial error")
	}

	// Echo loop.
	for {
		msg := <-conn.IncomingMessages()
		if msg.Data == nil {
			l.Debug().Msg("connection closed")
			break
		}

		l = l.With().Str("opcode", msg.Opcode.String()).Logger()
		l.Info().Int("length", len(msg.Data)).Msg("received message")

		switch msg.Opcode {
		case websocket.OpcodeText:
			err = <-conn.SendTextMessage(msg.Data)
		case websocket.OpcodeBinary:
			err = <-conn.SendBinaryMessage(msg.Data)
		default:
			l.Fatal().Msg("unexpected opcode in data message")
		}

		if err != nil {
			l.Err(err).Msg("echo error")
			conn.Close(websocket.StatusNormalClosure)
		}
	}
}
