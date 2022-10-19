package session

import (
	"time"
)

type timeout struct {
	start    time.Time
	timer    *time.Timer
	duration time.Duration
}

// String implements the fmt.Stringer interface,
// it returns the session time remaining rounded in seconds.
func (t *timeout) String() string {
	timeLeft := t.duration - time.Since(t.start)
	return "Time left: " + timeLeft.Round(time.Second).String()
}

// add adds time to the timeout.
func (t *timeout) add(d time.Duration) {
	if d == 0 {
		return
	}

	if t.duration == 0 {
		t.start = time.Now()
	}

	t.timer.Reset(t.duration - time.Since(t.start) + d)
	t.duration += d
}

// set sets the timeout time.
func (t *timeout) set(d time.Duration) {
	t.start = time.Now()
	t.duration = d

	if d == 0 {
		t.timer.Stop()
		return
	}

	t.timer.Reset(d)
}
