// Package timesim gives the rest of the project a FAKE clock instead of the
// real one.
//
// First, clear up a common wrong guess:
// You might think "my tests wait on real timers and that's slow, so I'll just
// use tiny real timeouts — sleep 10 milliseconds instead of 10 minutes." That
// does NOT fix the real problem. A 10ms sleep still goes through the operating
// system, which wakes it up a tiny bit differently each run — so two "identical"
// runs can still end up in a different order. And you still can't squeeze an
// hour-long scenario into a fast test; you're just sleeping a shorter real time.
//
// A fake clock is a totally different idea: fake time does NOT pass on its own
// at all. It only moves when something explicitly says "move forward" (see
// Advance in clock.go). So there's no waiting, and nothing for the OS to jitter.
//
// Why we need this:
// Go's real clock, time.Now(), never gives the same answer twice. A simulation
// that must be perfectly repeatable can't depend on it. So instead of calling
// time.Now() anywhere inside the simulated world, every package asks a Clock
// "what time is it?" — and this package is what answers, in a way the caller
// fully controls.
//
// What's inside:
//   - Clock    — a tiny interface: just "what time is it?" (see clock.go)
//   - SimClock — the fake clock everyone uses; starts where you tell it and only
//     moves when you call Advance
//
// How it's used: the scheduler owns ONE SimClock for the whole run. As it runs
// events in order, it moves the clock forward to each event's time before
// running it:
//
//	Scheduler
//	    │  Advance(d) — only the scheduler moves the clock
//	    ▼
//	SimClock
//	    │  Now() — anyone in the simulation may read it
//	    ▼
//	everything else in the simulated world
//
// One writer, many readers — that's what makes "what time is it?" a calm,
// repeatable question instead of a race.
//
// What goes wrong without it:
// Say a retry function measures "how long since I last tried?" using time.Now()
// directly. On a fast run the loop finishes in 2 real milliseconds; on a busy
// machine it takes 5. Same seed, same code — but the elapsed time it computes is
// different, so the outcome is different. One function quietly asked the real
// world, and repeatability broke.
//
// Important:
//
//   - Do NOT call time.Now() in this package, or in any package that takes a
//     Clock. One stray time.Now() brings back exactly the real-world jitter
//     above and breaks repeatability for the whole simulation.
//
// (Determinism note: a SimClock never moves on its own. Two runs that call
// Advance with the same durations, in the same order, see the exact same times
// come out of Now() — on any machine, every time.)
//
// One sentence to remember: this package isn't making time go faster — it's
// making time not pass on its own at all.
package timesim
