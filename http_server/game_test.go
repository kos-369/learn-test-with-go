package poker_test

import (
	"io"
	"testing"
	"time"

	poker "github.com/kos-369/learn-test-with-go/http_server"
)

type signalWriter chan struct{}

func (s signalWriter) Write(p []byte) (int, error) {
	s <- struct{}{}
	return len(p), nil
}

func TestTexasHoldemStop(t *testing.T) {
	alerter := &poker.SpyBlindAlerter{}
	game := poker.NewGame(alerter, dummyPlayerStore)

	stop := game.Start(5, io.Discard)

	if alerter.CancelCalls != 0 {
		t.Errorf("got %d alerts cancelled before stop, want 0", alerter.CancelCalls)
	}

	stop()

	if alerter.CancelCalls != len(alerter.Alerts) {
		t.Errorf("got %d alerts cancelled, want %d", alerter.CancelCalls, len(alerter.Alerts))
	}
}

func TestAlerter(t *testing.T) {
	t.Run("it writes the alert when it is not cancelled", func(t *testing.T) {
		fired := make(signalWriter, 1)

		poker.Alerter(time.Millisecond, 100, fired)

		select {
		case <-fired:
		case <-time.After(time.Second):
			t.Error("expected the alert to be written")
		}
	})

	t.Run("it does not write the alert after it is cancelled", func(t *testing.T) {
		fired := make(signalWriter, 1)

		cancel := poker.Alerter(5*tenMS, 100, fired)
		cancel()

		select {
		case <-fired:
			t.Error("expected the alert not to be written")
		case <-time.After(10 * tenMS):
		}
	})
}
