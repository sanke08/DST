package timesim

import "time"

// Clock is anything that can answer "what time is it right now?"
//
// Why it's an interface (not just the real clock type):
// If every package called time.Now() directly, we could never control time
// during a simulation. Instead, a package asks a Clock. Because it's an
// interface, the package doesn't know or care whether it's the real clock or a
// fake one — in this project it's always the fake one (SimClock, below).
//
// Think of Clock like a wall clock in a room: anyone can look up and read it
// (Now()), but only one person — the scheduler — is allowed to walk over and
// move the hands.
//
//	clock.Now() => whatever virtual time the clock currently shows
type Clock interface {
	Now() time.Time
}

// SimClock is the fake clock used everywhere in this project instead of the
// real one.
//
// It's just a stored time value that sits perfectly still until someone calls
// Advance on it. There's no background timer and no link to the operating
// system's clock — it never moves on its own.
//
// Important:
//
//   - Do NOT start a SimClock at time.Now() in code that must be repeatable.
//     If you did, two "identical" runs would start at two different real
//     moments, and if that start time ever gets printed or used as a key, the
//     two runs would disagree. Always pass a fixed start time.
type SimClock struct {
	// now is the clock's current virtual time. Only Advance changes it.
	now time.Time
}

// NewSimClock makes a SimClock that starts at the time you give it.
//
// There's no default and no version that quietly uses time.Now() — you must
// pick the starting time on purpose. That's the whole point of the package.
//
// Think of it like setting a stopwatch to a chosen starting number before a
// race, instead of trusting whatever time your phone happens to show:
//
//	NewSimClock(time.Unix(0, 0)) => a clock that starts at virtual time zero
func NewSimClock(start time.Time) *SimClock {
	return &SimClock{now: start}
}

// Now returns the clock's current virtual time.
//
//	NewSimClock(t).Now() => t, and it stays t until you call Advance
func (c *SimClock) Now() time.Time {
	return c.now
}

// Advance moves the clock forward by d. This is the ONLY thing that changes
// what Now() returns.
//
// The scheduler calls this once per event, moving the clock to that event's
// time before running it — so "what time is it?" always matches "which event
// are we running?"
//
// Think of it like moving a board-game piece forward by however many spaces the
// dice say — only forward, never back:
//
//	clock.Advance(5 * time.Second) => Now() now reports 5 seconds later
//
// Why it refuses to go backward:
// Say the clock reads 10:00:05 and, because of a bug, something asks to advance
// by -2 seconds. Code that already ran "later" would suddenly see time move
// backward — a message delivered later would look like it happened before an
// earlier one. That breaks the one thing this whole project promises: events
// happen in a fixed, sensible order. So a negative jump can only be a bug in
// the scheduler, not a normal thing to quietly allow.
//
// Important:
//
//   - Do NOT call Advance with a negative duration. It panics on purpose — a
//     negative jump means the scheduler computed an impossible time, which is a
//     bug to fix, not an error to recover from.
//
// (Determinism note: same start time + same list of durations, in the same
// order, always gives the same times out of Now() — on any machine.)
func (c *SimClock) Advance(d time.Duration) {
	if d < 0 {
		panic("timesim: Advance called with a negative duration; virtual time must never move backward")
	}
	c.now = c.now.Add(d)
}
