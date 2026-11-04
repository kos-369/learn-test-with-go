package poker

import (
	"fmt"
	"io"
	"time"
)

type BlindAlerter interface {
	ScheduleAlertAt(duration time.Duration, amount int, to io.Writer) (cancel func())
}

type BlindAlerterFunc func(duration time.Duration, amount int, to io.Writer) (cancel func())

func (a BlindAlerterFunc) ScheduleAlertAt(duration time.Duration, amount int, to io.Writer) (cancel func()) {
	return a(duration, amount, to)
}

func Alerter(duration time.Duration, amount int, to io.Writer) (cancel func()) {
	timer := time.AfterFunc(duration, func() {
		fmt.Fprintf(to, "Blind is now %d\n", amount)
	})

	return func() {
		timer.Stop()
	}
}
