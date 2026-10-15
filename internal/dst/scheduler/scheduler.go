package scheduler

import (
	"container/heap"
	"time"

	"github.com/sanke08/DST/internal/dst/core"
	"github.com/sanke08/DST/internal/dst/randsim"
	"github.com/sanke08/DST/internal/dst/timesim"
)

// Scheduler is the engine that actually runs a simulation. It is the one thing
// that turns the passive pieces you've built — the SimClock, the seeded RNG, and
// the Event queue — into a run that plays out in a fixed, repeatable order.
//
// The mental model — a to-do list with times on it:
// Everything the simulation will ever do is written down as an Event with a
// virtual FireTime ("deliver this message at time 100"). The Scheduler keeps
// that to-do list sorted by time, and works through it strictly one item at a
// time: take the earliest, move the clock to it, do it, repeat. Nothing runs in
// the background; nothing runs "whenever." That single-threaded, earliest-first
// loop is the whole reason a run is repeatable.
//
// Everything a run needs to be deterministic lives INSIDE one Scheduler:
//   - its own event queue
//   - its own SimClock (virtual time)
//   - its own seeded RNG
//   - its own ID/seq counters
//
// So there is no shared global state between runs. Build a Scheduler from a
// seed + start time, feed it the same setup, and you get the identical run.
type Scheduler struct {
	// queue is the sorted waiting line of events not yet run. Managed only
	// through container/heap so the earliest-first + seq-tiebreak order holds.
	queue eventQueue

	// clock is this run's virtual time. The Scheduler is the ONLY thing allowed
	// to move it (via clock.Advance), and it only ever moves it forward to the
	// FireTime of the event about to run — see Step.
	clock *timesim.SimClock

	// rng is this run's randomness, seeded once at construction. Actions pull
	// their "random" delays/drops from here (via Rand()) so the randomness is
	// part of the same repeatable stream, not a separate uncontrolled source.
	rng *randsim.SimRNG

	// nextID / nextSeq are plain counters, both starting at 1. nextID names each
	// event (EventID); nextSeq records the order events were scheduled so ties on
	// FireTime break the same way every run. They only ever go up, one per
	// Schedule call. Starting at 1 keeps 0 meaning "not a real event" (the same 0
	// rule as core.EventID).
	nextID  uint64
	nextSeq uint64
}

// NewScheduler builds a fresh engine for one run.
//
// You must supply BOTH inputs on purpose — there is no version that reads the
// wall clock or picks a random seed for you, because that would make the run
// unrepeatable:
//
//   - seed decides the entire random stream (passed straight to NewRNG).
//   - start is the virtual time the clock begins at (passed to NewSimClock).
//
// Everything else (queue, counters) starts empty/at 1.
func NewScheduler(seed core.Seed, start time.Time) *Scheduler {
	return &Scheduler{
		queue:   eventQueue{},
		clock:   timesim.NewSimClock(start),
		rng:     randsim.NewRNG(seed),
		nextID:  1,
		nextSeq: 1,
	}
}

// Now reports the current virtual time. It's a read-only window onto the clock
// for code that has a Scheduler but shouldn't be able to move time itself.
func (s *Scheduler) Now() time.Time {
	return s.clock.Now()
}

// Rand hands back this run's random-number source. Actions that need a random
// delay or a chance-based decision should draw from HERE, not from their own
// generator, so all randomness stays part of the one seeded stream.
func (s *Scheduler) Rand() randsim.RNG {
	return s.rng
}

// Len reports how many events are still waiting to run. Handy for a caller that
// wants to loop "while there's work left" by hand instead of calling Run.
func (s *Scheduler) Len() int {
	return s.queue.Len()
}

// Schedule adds a new event to run `delay` from the current virtual time, and
// returns the EventID naming it (so a future feature could, say, cancel it).
//
// What it does, step by step:
//   - assigns the next EventID and the next seq (the ordering tiebreak)
//   - computes FireTime = now + delay
//   - pushes the event into the sorted queue
//
// Two important rules:
//   - delay must not be negative. A negative delay would place an event in the
//     past, which the earliest-first loop can't honor sensibly — that's a bug in
//     the caller, so we panic rather than silently run it "now."
//   - action must not be nil. An event with nothing to do is meaningless and
//     would panic later when Step tries to run it; catching it here points at the
//     real mistake.
//
// Scheduling from INSIDE a running action is fine and expected (a message
// handler scheduling a reply). The new event just lands in the queue at its
// FireTime and gets picked up in order like any other.
func (s *Scheduler) Schedule(delay time.Duration, action func()) core.EventID {
	if delay < 0 {
		panic("scheduler: Schedule called with a negative delay; events cannot run in the past")
	}
	if action == nil {
		panic("scheduler: Schedule called with a nil action; an event must have something to do")
	}

	ev := &Event{
		ID:       core.EventID(s.nextID),
		FireTime: s.clock.Now().Add(delay),
		seq:      s.nextSeq,
		Action:   action,
	}
	s.nextID++
	s.nextSeq++

	heap.Push(&s.queue, ev)
	return ev.ID
}

// Step runs exactly ONE event — the earliest one waiting — and returns true. If
// the queue is empty it does nothing and returns false, which is how a hand-
// written loop knows to stop.
//
// The order here is the heart of determinism:
//  1. pop the earliest event (heap gives it back using FireTime, then seq)
//  2. advance the clock forward TO that event's FireTime — so "what time is it?"
//     always matches "which event are we running?"
//  3. run its Action
//
// Because the clock only ever moves to the event we're about to run, an Action
// that schedules more work always schedules it at a time >= now, so time never
// needs to go backward (SimClock.Advance would panic if it did).
func (s *Scheduler) Step() bool {
	if s.queue.Len() == 0 {
		return false
	}

	ev := heap.Pop(&s.queue).(*Event)
	s.clock.Advance(ev.FireTime.Sub(s.clock.Now()))
	ev.Action()
	return true
}

// Run works through the entire queue until nothing is left, running events
// earliest-first, one at a time. It's the normal way to play out a simulation:
// schedule some starting events, then call Run.
//
// Note that Run naturally handles events that schedule MORE events: those get
// added mid-run and Run keeps going until the queue truly drains. If a
// simulation schedules work forever, Run runs forever — that's on the caller's
// setup, not something Run guesses about.
func (s *Scheduler) Run() {
	for s.Step() {
	}
}