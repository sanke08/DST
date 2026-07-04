package scheduler

import (
	"time"

	"github.com/sanke08/DST/internal/dst/core"
)

// Event is one thing the simulation has scheduled to happen at a specific
// virtual moment — "deliver this message at virtual-time 100," "fire this
// retry timer at virtual-time 250," and so on.
//
// Why this exists:
// The scheduler (see scheduler.go) doesn't run the simulation as a stream of
// live function calls happening whenever code feels like calling them. Instead
// everything that will happen "later" is turned into one of these Event values
// and dropped into a queue, so the scheduler can decide the exact order the
// whole run unfolds in. Making "a future action" a plain, inspectable value —
// rather than, say, a goroutine sleeping until its time comes — is what lets
// the run be paused, ordered, and reproduced. A goroutine's wake-up time is at
// the mercy of the OS scheduler; an Event sitting in a queue is entirely under
// this framework's control.
//
// What it does:
// Bundles four things: which event this is (ID), when it should fire
// (FireTime), a tiebreaker for events that fire at the exact same instant
// (seq — explained below and used by the queue in this same file), and what to
// actually do when it fires (Action).
//
// Determinism note:
// Nothing about an Event is derived from wall-clock time or randomness. Its
// FireTime is a virtual time the scheduler computed; its ID and seq are
// assigned from plain counters. So two runs that schedule the same events in
// the same order produce Events that compare and order identically — which is
// the property the queue below relies on.
type Event struct {
	// ID names this specific event so it can be referred to later — for
	// instance, so a trace (Phase 5) can record "event-42 fired at virtual
	// time 100," or so a future cancel-a-pending-timer feature can point at
	// exactly one event. It's assigned by the scheduler from an incrementing
	// counter, never from anything random or time-based.
	ID core.EventID

	// FireTime is the virtual moment this event should run — not a real
	// wall-clock time, but a point on the SimClock the scheduler owns. The
	// queue below orders events primarily by this field: earliest FireTime
	// runs first.
	//
	// Important:
	//
	//   - Do NOT compare FireTime values across two different simulation
	//     runs and expect them to line up with any real-world clock. A
	//     FireTime is only meaningful relative to the SimClock's start time
	//     that this run was constructed with.
	FireTime time.Time

	// seq is a monotonically increasing sequence number — 1, 2, 3, ... in the
	// exact order events were scheduled — used purely as a tiebreaker.
	//
	// Why this field is the whole point of the type:
	// Two events can legitimately share the exact same FireTime (schedule two
	// messages "5 seconds from now" back to back and both land on the same
	// virtual instant). Something has to decide which of two same-time events
	// runs first, and that decision must be identical on every run or the
	// whole framework's promise collapses. seq is that decider: among events
	// tied on FireTime, the one scheduled first (smaller seq) runs first.
	//
	// What goes wrong without it — the failure this field prevents:
	// If ties were broken by anything ambient — the order two entries happen
	// to sit in a Go map (deliberately randomized by the runtime), a pointer
	// address, or which of two goroutines got there first — then two runs of
	// the same seed could process the same two same-time events in opposite
	// orders. One run delivers message A then B; the next delivers B then A;
	// and a bug that depends on that order is un-reproducible, which is the
	// exact disease this whole project exists to cure. Pinning the tiebreak to
	// a deterministic counter turns "whatever the runtime felt like" into "the
	// order they were scheduled," every time.
	//
	// Why lowercase (unexported):
	// This is bookkeeping the scheduler owns; nothing outside this package
	// should set or read it. Exporting it would invite a caller to fiddle with
	// tie ordering, which is precisely the ambient influence this field exists
	// to eliminate.
	seq uint64

	// Action is what actually happens when this event fires — deliver a
	// message into a node, fire a timer callback, apply a fault. Per this
	// project's Phase 0 decision (single-threaded, callback-driven simulation),
	// the scheduler runs these one at a time, never concurrently, so an Action
	// never has to worry about another Action running at the same moment.
	//
	// Why a plain func, and the one honest limitation of that choice:
	// A closure is the simplest, most idiomatic way to say "run this later,"
	// and it lets Schedule callers capture whatever they need inline. The
	// tradeoff, called out plainly rather than hidden: a func value cannot be
	// serialized. So when the scheduler is snapshotted (snapshot.go, later this
	// phase), the Action is NOT part of what round-trips to disk — the queue's
	// timing/ordering metadata is saved, but the closures are re-supplied by
	// re-running the same setup code. That's not a defect; it's exactly how
	// replay-from-seed is meant to work: run the identical program against the
	// identical seed and you get the identical closures back for free.
	Action func()
}
