package scheduler

import (
	"time"

	"github.com/sanke08/DST/internal/dst/core"
)

// Event is one thing the simulation has planned to do at a specific virtual
// time — for example "deliver this message at time 100" or "fire this retry
// timer at time 250."
//
// Why turn a future action into a stored value like this?
// The scheduler does NOT run the simulation by just calling functions whenever
// code feels like it. Instead, everything that should happen "later" is packed
// into one of these Event values and put in a queue. That lets the scheduler
// decide the exact order the whole run plays out in.
//
// Compare it to two ways of setting a reminder:
//   - A goroutine that sleeps until its time comes is like telling a friend
//     "wake me in 5 minutes" — the exact moment they tap you depends on them,
//     and you can't control it.
//   - An Event in a queue is like writing "9:05 — call mom" on a list. It just
//     sits there as plain, readable data until the scheduler decides to run it.
//     That control is what makes the run repeatable.
//
// (Determinism note: nothing in an Event comes from the real clock or from
// randomness. Its time is a virtual time the scheduler computed, and its ID and
// seq are plain counters. So two runs that schedule the same events in the same
// order produce Events that sort the same way every time.)
type Event struct {
	// ID names this one event, so it can be pointed at later — for example, so
	// a future "cancel that pending timer" feature can say WHICH timer, or a
	// trace can record "event-42 ran at time 100." Handed out by the scheduler
	// from a simple counter, never anything random.
	ID core.EventID

	// FireTime is the virtual moment this event should run — a point on the
	// fake SimClock, NOT a real wall-clock time. The queue runs events in
	// FireTime order: earliest time first.
	//
	// Important:
	//
	//   - Do NOT compare a FireTime from one run against a real clock or against
	//     another run's times. It only means something relative to the start
	//     time this particular run's clock was set to.
	FireTime time.Time

	// seq is a plain counter — 1, 2, 3, ... in the order events were scheduled.
	// Its only job is to break TIES, and it's the reason this whole type is
	// careful.
	//
	// The problem it solves:
	// Two events can have the EXACT same FireTime (schedule two messages "5
	// seconds from now" one after the other — both land on the same virtual
	// instant). Something has to decide which of the two runs first, and that
	// decision must be the same on every run.
	//
	// What would go wrong without it:
	// If we broke ties using anything that changes run-to-run — the order two
	// keys happen to sit in a Go map (Go deliberately randomizes that!), a
	// memory address, whichever goroutine got there first — then one run would
	// deliver message A then B, and the next run B then A. A bug that depends on
	// that order becomes impossible to reproduce, which is the exact disease
	// this whole project is trying to cure.
	//
	// The fix is simple: among events tied on FireTime, the one scheduled first
	// (smaller seq) runs first. Always the same, every run.
	//
	// It's lowercase (private) because only the scheduler should set or read it.
	// Letting outside code touch it would reopen the very "unpredictable order"
	// hole it exists to close.
	seq uint64

	// Action is what actually happens when this event runs — deliver a message,
	// fire a timer, apply a fault. The scheduler runs these one at a time, never
	// two at once, so an Action never has to worry about another running
	// alongside it.
	//
	// One honest limitation of using a plain function here:
	// A function value can't be saved to disk. So when the scheduler's state is
	// snapshotted later, the Action is NOT part of what gets saved — only the
	// timing/order info is. That's fine and by design: to replay a run you
	// re-run the same setup code with the same seed, which hands you back the
	// exact same functions for free.
	Action func()
}
