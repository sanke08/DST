package timesim

import "time"

// Clock is anything that can answer "what time is it right now?"
//
// Why this exists:
// If every package that needs "the current time" just called time.Now()
// directly, there would be no way to control time during a simulation —
// each call would read the real, ever-changing wall clock. By instead
// taking a Clock as a constructor argument, a package doesn't know or care
// whether it's talking to the real clock or a fully controllable fake one.
// In this project it's always the fake one (SimClock, below) — nothing here
// ever wires up the real clock.
//
// What it does:
// A one-method interface: Now() returns the current virtual time. That's
// the entire contract. Anything that implements this one method can stand
// in for "the clock" anywhere in this project.
//
// Example:
// Think of Clock like a wall clock hanging in a room. Anyone in the room
// can look up and read it (`clock.Now()`), but only one person — the
// scheduler, in this project — is allowed to walk over and move the hands.
//
//	clock.Now() ⇒ whatever virtual time the clock currently shows.
type Clock interface {
	Now() time.Time
}

// SimClock is the simulated clock used everywhere in this project instead
// of the real one.
//
// Why this exists:
// This is the concrete type that makes Clock deterministic: a SimClock only
// ever changes when something calls Advance on it directly (see Advance).
// There's no background goroutine, no tie to the operating system's clock —
// it's just a stored time.Time value that sits still until told to move.
//
// What it does:
// Holds one virtual "current time" value. Now() reads it; Advance (below)
// is the only thing that moves it forward.
//
// Important:
//
//   - Do NOT construct a SimClock with time.Now() as the start time in any
//     code that's meant to be deterministic (anything other than a one-off
//     manual script). Doing so would make two "identical" runs start at two
//     different real-world moments, and if that start time is ever printed,
//     logged, or used as a key, two runs of the same seed would disagree.
//     Always pass an explicit, fixed start time.
type SimClock struct {
	now time.Time
}

// NewSimClock creates a SimClock that starts at the given time.
//
// Why this exists:
// The caller must decide the starting virtual time explicitly — that's the
// whole point of this package. There's no default, and no version of this
// constructor that reaches for time.Now() on the caller's behalf.
//
// Example:
// Think of it like setting a stopwatch to a specific starting number before
// a race, instead of trusting whatever time your phone happens to show.
//
//	NewSimClock(time.Unix(0, 0)) ⇒ a clock that starts at virtual time zero.
func NewSimClock(start time.Time) *SimClock {
	return &SimClock{now: start}
}

// Now returns the clock's current virtual time.
//
// Example: a freshly constructed NewSimClock(t).Now() ⇒ t, unchanged until
// Advance is called.
func (c *SimClock) Now() time.Time {
	return c.now
}

// Advance moves the clock's virtual time forward by d.
//
// Why this exists:
// This is the only method that changes what Now() returns. The scheduler
// (Phase 3) will call this once per processed event, advancing the clock to
// that event's virtual timestamp before running it — so "what time is it"
// always matches "what event are we currently processing."
//
// Example:
// Think of it like moving a board-game piece forward by however many spaces
// the dice say — never backward, and never more than one turn's worth at a
// time.
//
//	clock.Advance(5 * time.Second) ⇒ Now() reports 5 seconds later than before.
//
// What goes wrong without the negative-duration guard:
// Say the clock reads 10:00:05 and the scheduler is about to process two
// pending events, one timestamped 10:00:10 and one timestamped 10:00:08.
// If the scheduler's own bookkeeping had a bug and asked to advance by -2
// seconds partway through (going from 10:00:10 back to 10:00:08), any code
// that had already run "at 10:00:10" and read the clock would see time run
// backward — a message that was supposedly delivered later would appear to
// exist before one delivered earlier. That's not a rare edge case to
// tolerate; it's a direct contradiction of "events happen in a fixed
// order," which is the one thing this whole framework promises.
//
// Important:
//
//   - Do NOT call Advance with a negative duration. Advance panics if d is
//     negative, on purpose: a negative duration here means the scheduler
//     itself computed an impossible timestamp, which is a bug in this
//     framework's core, not a normal error a caller should be expected to
//     recover from.
//
// Determinism note:
// Calling Advance with the same starting time and the same sequence of
// durations always produces the same sequence of Now() results, on any
// machine. That's the entire determinism guarantee this type provides.
func (c *SimClock) Advance(d time.Duration) {
	if d < 0 {
		panic("timesim: Advance called with a negative duration; virtual time must never move backward")
	}
	c.now = c.now.Add(d)
}
