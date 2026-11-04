package poker_test

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	poker "github.com/kos-369/learn-test-with-go/http_server"
)

var dummyStdOut = &bytes.Buffer{}
var dummyPlayerStore = &poker.StubPlayerStore{}

const tenMS = 10 * time.Millisecond

func userSends(messages ...string) io.Reader {
	return strings.NewReader(strings.Join(messages, "\n"))
}

func TestCLI(t *testing.T) {
	t.Run("start game with 3 players and finish game with 'Chris' as winner", func(t *testing.T) {
		game := &poker.GameSpy{}

		out := &bytes.Buffer{}
		in := userSends("3", "Chris wins")

		poker.NewCLI(in, out, game).PlayPoker()

		assertMessagesSentToUser(t, out, poker.PlayerPrompt)
		assertGameStartedWith(t, game, 3)
		assertFinishCalledWith(t, game, "Chris")
	})

	t.Run("start game with 8 players and record 'Cleo' as winner", func(t *testing.T) {
		game := &poker.GameSpy{}

		in := userSends("8", "Cleo wins")
		cli := poker.NewCLI(in, dummyStdOut, game)

		cli.PlayPoker()

		assertGameStartedWith(t, game, 8)
		assertFinishCalledWith(t, game, "Cleo")
	})

	t.Run("it prints an error when a non numeric value is entered and does not start the game", func(t *testing.T) {
		game := &poker.GameSpy{}

		stdout := &bytes.Buffer{}
		in := userSends("pies")

		cli := poker.NewCLI(in, stdout, game)
		cli.PlayPoker()

		assertGameNotStarted(t, game)
		assertMessagesSentToUser(t, stdout, poker.PlayerPrompt, poker.BadPlayerInputErrMsg)
	})

	t.Run("start a game with 3 players, send some blind alerts down WS and declare Ruth the winner", func(t *testing.T) {
		wantedBlindAlert := "Blind is 100"
		winner := "Ruth"

		game := &poker.GameSpy{BlindAlert: []byte(wantedBlindAlert)}
		server := httptest.NewServer(mustMakePlayerServer(t, dummyPlayerStore, game))
		ws := mustDialWS(t, "ws"+strings.TrimPrefix(server.URL, "http")+"/ws")

		defer server.Close()
		defer ws.Close()

		writeWSMessage(t, ws, "3")
		writeWSMessage(t, ws, winner)

		time.Sleep(tenMS)

		assertGameStartedWith(t, game, 3)
		assertFinishCalledWith(t, game, winner)
		within(t, tenMS, func() { assertWebsocketGotMsg(t, ws, wantedBlindAlert) })
	})

	for _, input := range []string{"0", "-5"} {
		t.Run("it does not start the game with "+input+" players", func(t *testing.T) {
			game := &poker.GameSpy{}

			stdout := &bytes.Buffer{}
			in := userSends(input)

			poker.NewCLI(in, stdout, game).PlayPoker()

			assertGameNotStarted(t, game)
			assertMessagesSentToUser(t, stdout, poker.PlayerPrompt, poker.BadPlayerInputErrMsg)
		})
	}

	inputsWithoutWinner := map[string]io.Reader{
		"input ends before a winner is entered": userSends("3"),
		"a blank winner is entered":             userSends("3", "   "),
	}

	for name, in := range inputsWithoutWinner {
		t.Run("it does not record a winner when "+name, func(t *testing.T) {
			game := &poker.GameSpy{}

			stdout := &bytes.Buffer{}

			poker.NewCLI(in, stdout, game).PlayPoker()

			assertGameStartedWith(t, game, 3)
			assertGameNotFinished(t, game)
			assertGameStopped(t, game)
			assertMessagesSentToUser(t, stdout, poker.PlayerPrompt, poker.BadWinnerInputErrMsg)
		})
	}

	t.Run("it stops the blind alerts when the game finishes", func(t *testing.T) {
		game := &poker.GameSpy{}

		poker.NewCLI(userSends("3", "Chris wins"), dummyStdOut, game).PlayPoker()

		assertFinishCalledWith(t, game, "Chris")
		assertGameStopped(t, game)
	})
}

func TestWebSocketGame(t *testing.T) {
	startServer := func(t *testing.T, game poker.Game) *websocket.Conn {
		t.Helper()
		server := httptest.NewServer(mustMakePlayerServer(t, dummyPlayerStore, game))
		t.Cleanup(server.Close)

		ws := mustDialWS(t, "ws"+strings.TrimPrefix(server.URL, "http")+"/ws")
		t.Cleanup(func() { ws.Close() })

		return ws
	}

	for _, input := range []string{"abc", "", "0", "-5"} {
		t.Run("it does not start the game with players "+input, func(t *testing.T) {
			game := &poker.GameSpy{}
			ws := startServer(t, game)

			writeWSMessage(t, ws, input)

			assertServerClosedConnection(t, ws)
			assertGameNotStarted(t, game)
		})
	}

	for _, winner := range []string{"", "   "} {
		t.Run("it does not record the blank winner "+winner, func(t *testing.T) {
			game := &poker.GameSpy{}
			ws := startServer(t, game)

			writeWSMessage(t, ws, "3")
			writeWSMessage(t, ws, winner)

			assertServerClosedConnection(t, ws)
			assertGameStartedWith(t, game, 3)
			assertGameNotFinished(t, game)
			assertGameStopped(t, game)
		})
	}

	t.Run("it stops the game without recording a winner when the client disconnects", func(t *testing.T) {
		game := &poker.GameSpy{}
		ws := startServer(t, game)

		writeWSMessage(t, ws, "3")
		assertGameStartedWith(t, game, 3)

		ws.Close()

		assertGameStopped(t, game)
		assertGameNotFinished(t, game)
	})

	t.Run("it closes the connection and stops the game after the winner is declared", func(t *testing.T) {
		game := &poker.GameSpy{}
		ws := startServer(t, game)

		writeWSMessage(t, ws, "3")
		writeWSMessage(t, ws, "Ruth")

		assertServerClosedConnection(t, ws)
		assertFinishCalledWith(t, game, "Ruth")
		assertGameStopped(t, game)
	})
}

func assertMessagesSentToUser(t testing.TB, stdout *bytes.Buffer, messages ...string) {
	t.Helper()
	want := strings.Join(messages, "")
	got := stdout.String()
	if got != want {
		t.Errorf("got %q sent to stdout but expected %+v", got, messages)
	}
}

func assertFinishCalledWith(t testing.TB, game *poker.GameSpy, winner string) {
	t.Helper()
	passed := retryUntil(500*time.Millisecond, func() bool {
		return game.FinishedWith() == winner
	})
	if !passed {
		t.Errorf("expected finish called with %q but got %q", winner, game.FinishedWith())
	}
}

func assertGameNotStarted(t testing.TB, game *poker.GameSpy) {
	t.Helper()
	if game.WasStarted() {
		t.Errorf("game should not have started")
	}
}

func assertGameStartedWith(t testing.TB, game *poker.GameSpy, numberOfPlayersWanted int) {
	t.Helper()
	passed := retryUntil(500*time.Millisecond, func() bool {
		return game.StartedWith() == numberOfPlayersWanted
	})
	if !passed {
		t.Errorf("wanted Start called with %d but got %d", numberOfPlayersWanted, game.StartedWith())
	}
}

func assertGameNotFinished(t testing.TB, game *poker.GameSpy) {
	t.Helper()
	if game.WasFinished() {
		t.Errorf("game should not have finished")
	}
}

func assertGameStopped(t testing.TB, game *poker.GameSpy) {
	t.Helper()
	passed := retryUntil(500*time.Millisecond, game.WasStopped)
	if !passed {
		t.Errorf("expected the game to be stopped")
	}
}

func assertServerClosedConnection(t testing.TB, ws *websocket.Conn) {
	t.Helper()
	ws.SetReadDeadline(time.Now().Add(time.Second))

	for {
		_, _, err := ws.ReadMessage()
		if err == nil {
			continue
		}

		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			t.Fatal("server did not close the connection")
		}

		return
	}
}

func retryUntil(d time.Duration, f func() bool) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if f() {
			return true
		}
	}
	return false
}

func mustMakePlayerServer(t *testing.T, store poker.PlayerStore, game poker.Game) *poker.PlayerServer {
	t.Helper()
	server, err := poker.NewPlayerServer(store, game)
	if err != nil {
		t.Fatal("problem creating player server", err)
	}
	return server
}

func mustDialWS(t *testing.T, url string) *websocket.Conn {
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", url, err)
	}

	return ws
}

func writeWSMessage(t testing.TB, conn *websocket.Conn, message string) {
	t.Helper()
	if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
		t.Fatalf("could not send message over ws connection %v", err)
	}
}

func within(t testing.TB, d time.Duration, assert func()) {
	t.Helper()

	done := make(chan struct{}, 1)

	go func() {
		assert()
		done <- struct{}{}
	}()

	select {
	case <-time.After(d):
		t.Error("timed out")
	case <-done:
	}
}

func assertWebsocketGotMsg(t testing.TB, ws *websocket.Conn, want string) {
	t.Helper()
	_, msg, err := ws.ReadMessage()
	if err != nil {
		t.Errorf("could not read message from ws connection %v", err)
		return
	}
	if string(msg) != want {
		t.Errorf(`got "%s", want "%s"`, string(msg), want)
	}
}
