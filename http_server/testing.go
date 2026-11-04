package poker

import (
	"fmt"
	"io"
	"sync"
	"testing"
	"time"
)

type StubPlayerStore struct {
	Scores   map[string]int
	WinCalls []string
	League   []Player
}

func (s *StubPlayerStore) GetPlayerScore(name string) int {
	score := s.Scores[name]
	return score
}

func (s *StubPlayerStore) RecordWin(name string) {
	s.WinCalls = append(s.WinCalls, name)
}

func (s *StubPlayerStore) GetLeague() League {
	return s.League
}

func AssertPlayerWin(t testing.TB, store *StubPlayerStore, winner string) {
	t.Helper()

	if len(store.WinCalls) != 1 {
		t.Fatalf("got %d calls to RecordWin want %d", len(store.WinCalls), 1)
	}

	if store.WinCalls[0] != winner {
		t.Errorf("did not store correct winner got %q want %q", store.WinCalls[0], winner)
	}
}

type ScheduledAlert struct {
	At     time.Duration
	Amount int
}

func (s ScheduledAlert) String() string {
	return fmt.Sprintf("%d chips at %v", s.Amount, s.At)
}

type SpyBlindAlerter struct {
	Alerts      []ScheduledAlert
	CancelCalls int
}

func (s *SpyBlindAlerter) ScheduleAlertAt(at time.Duration, amount int, to io.Writer) (cancel func()) {
	s.Alerts = append(s.Alerts, ScheduledAlert{at, amount})

	return func() {
		s.CancelCalls++
	}
}

type GameSpy struct {
	mu sync.Mutex

	StartCalled     bool
	StartCalledWith int
	BlindAlert      []byte

	FinishedCalled   bool
	FinishCalledWith string

	Stopped bool
}

func (g *GameSpy) Start(numberOfPlayers int, out io.Writer) (stop func()) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.StartCalled = true
	g.StartCalledWith = numberOfPlayers
	out.Write(g.BlindAlert)

	return func() {
		g.mu.Lock()
		defer g.mu.Unlock()
		g.Stopped = true
	}
}

func (g *GameSpy) Finish(winner string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.FinishedCalled = true
	g.FinishCalledWith = winner
}

func (g *GameSpy) WasStarted() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.StartCalled
}

func (g *GameSpy) StartedWith() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.StartCalledWith
}

func (g *GameSpy) WasFinished() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.FinishedCalled
}

func (g *GameSpy) FinishedWith() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.FinishCalledWith
}

func (g *GameSpy) WasStopped() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.Stopped
}
