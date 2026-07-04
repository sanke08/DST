// Package timesim gives the rest of the DST framework a fake clock instead
// of the real one.
//
// A common misconception, cleared up first:
// The obvious-sounding fix for "tests that wait on real timers are slow" is
// to just use shorter real timeouts — sleep 10 milliseconds in a test
// instead of the 10 minutes a retry loop waits in production. That doesn't
// actually solve the problem this package exists for. Those 10ms sleeps
// still go through the real operating system scheduler, which wakes them
// up at a slightly different moment every run — so two "identical" test
// runs can still interleave events in a different order. And it still
// can't compress a genuinely long scenario (an hour of retries) into a
// fast test; you're just waiting a shorter real amount of time. A
// simulated clock solves a different problem entirely: virtual time
// doesn't pass on its own AT ALL. It only moves when something explicitly
// says "advance to the next event" (see Advance in clock.go) — so there's
// no waiting, real or otherwise, and nothing for the OS to jitter.
//
// Why this exists:
// Go's real clock — time.Now() — never gives the same answer twice; it's
// tied to the actual wall clock ticking on whatever machine runs the code.
// A simulation that's supposed to be perfectly repeatable (same seed ⇒ same
// run, forever) cannot have anything in it that depends on wall-clock time.
// So instead of calling time.Now() anywhere inside the simulated world, every
// package in this project takes a Clock as a constructor argument and asks
// it "what time is it?" — and this package is what answers that question in
// a way that's fully under the caller's control.
//
// What it does:
// Defines the Clock interface (just "what time is it right now?") and
// SimClock, the one implementation of it used everywhere in this project.
// A SimClock starts at a time the caller picks and only moves forward when
// something explicitly tells it to.
//
// How it works:
// The scheduler (Phase 3) will own one SimClock for an entire simulation
// run. As it processes scheduled events in order, it advances the SimClock
// to each event's virtual timestamp before running that event:
//
//	Scheduler
//	    │  Advance(d) — the only thing that moves the clock
//	    ▼
//	SimClock
//	    │  Now() — anyone in the simulation can read it
//	    ▼
//	everything else in the simulated world
//
// One writer, many readers — that's what makes "what time is it" a
// deterministic question instead of a race.
//
// What goes wrong without this:
// Imagine a retry-backoff function that computes "how long has it been
// since I last tried?" using time.Now() directly, instead of asking a
// Clock. On one run, the machine is under light load and the whole retry
// loop finishes in a few real milliseconds. On another run — same seed,
// same code — the machine is briefly busy, real milliseconds tick by
// differently between two time.Now() calls, and the backoff function
// computes a different elapsed duration. The seed never changed, but the
// simulation's outcome did, because one function quietly asked the real
// world instead of the fake clock everyone else was using.
//
// Important:
//
//   - Do NOT call time.Now() anywhere in this package, or in any package
//     that takes a Clock. A single stray time.Now() call reintroduces
//     exactly the real-world jitter described above, breaking
//     reproducibility for the entire simulation even if everything else is
//     done correctly.
//
// Determinism note:
// A SimClock never changes on its own — it only moves when a caller calls
// Advance (see clock.go). Two simulation runs that call Advance with the
// same sequence of durations will see the exact same sequence of Now()
// values, on any machine, every time.
//
// The one sentence to remember: this package isn't trying to make time
// pass faster — it's trying to make time not pass on its own at all.
package timesim
