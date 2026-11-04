package poker

import (
	"io"
	"time"
)

type TexasHoldem struct {
	alerter BlindAlerter
	store   PlayerStore
}

func NewGame(alerter BlindAlerter, store PlayerStore) *TexasHoldem {
	return &TexasHoldem{
		alerter: alerter,
		store:   store,
	}
}

func (p *TexasHoldem) Start(numberOfPlayers int, alertsDestination io.Writer) (stop func()) {
	blindIncrement := time.Duration(5+numberOfPlayers) * time.Minute

	blinds := []int{100, 200, 300, 400, 500, 600, 800, 1000, 2000, 4000, 8000}
	cancels := make([]func(), 0, len(blinds))
	blindTime := 0 * time.Second
	for _, blind := range blinds {
		cancels = append(cancels, p.alerter.ScheduleAlertAt(blindTime, blind, alertsDestination))
		blindTime = blindTime + blindIncrement
	}

	return func() {
		for _, cancel := range cancels {
			cancel()
		}
	}
}

func (p *TexasHoldem) Finish(winner string) {
	p.store.RecordWin(winner)
}
