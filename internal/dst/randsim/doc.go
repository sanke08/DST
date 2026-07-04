// Package randsim gives the rest of the DST framework a fake source of
// randomness instead of the real one.
//
// A common misconception, cleared up first:
// "Same seed" does NOT mean "same number every time you ask." A PRNG
// (pseudo-random number generator) hands back the next number in one long,
// fixed sequence — which number you get depends on how many times it's
// already been asked, not just on the seed. Seed 42 might always produce:
//
//	1st call ⇒ 83
//	2nd call ⇒ 17
//	3rd call ⇒ 56
//
// Every single time you start over from seed 42 and ask three times, you
// get 83, then 17, then 56, in that order — never three 83s in a row. Think
// of it like a shuffled deck: the same seed gives you the same order of
// cards, not the same card every draw.
//
// Why this exists:
// That "depends on call position" detail is exactly why this package can't
// just use Go's math/rand the normal way (rand.Intn, rand.Float64 — the
// global functions). Those all draw from one generator shared by the whole
// program. If two unrelated parts of the program both draw from it, each
// one's calls shift the other's position in the sequence. Concretely, say
// the simulation expects:
//
//	Scheduler asks (1st call) ⇒ 83
//	Scheduler asks (2nd call) ⇒ 17
//
// Now suppose a logging package — nothing to do with the simulation —
// also calls the global rand.Intn once, right in between:
//
//	Scheduler asks (1st call) ⇒ 83   — as expected
//	Logger asks    (2nd call) ⇒ 17   — the logger stole this one
//	Scheduler asks (3rd call) ⇒ 56   — the scheduler wanted 17, got 56
//
// The seed never changed. Nobody touched the simulation's code. Adding one
// unrelated log line silently changed the simulation's outcome, because it
// changed which call number the scheduler's second request landed on. A
// deterministic simulator cannot tolerate that — "same seed" has to mean
// "same outcome," full stop, regardless of what else is running.
//
// The shape of the problem, and the fix, side by side:
//
//	the broken, shared way                the way this package does it
//
//	   Program                                Simulation
//	  /   |    \                                  │
//	Sim  Logger Metrics                           ▼
//	  \   |    /                                SimRNG
//	      ▼                                       │
//	  Global RNG                                  ▼
//	                                          *rand.Rand
//
// On the left, three unrelated things pull from one shared sequence, so any
// one of them can shift what the others get. On the right, a SimRNG is
// wired to exactly one simulation and wraps its own private *rand.Rand —
// there's no third party positioned to steal a call from it, because
// nothing else in the program even has a reference to it.
//
// What it does:
// Defines the RNG interface (a small set of "give me a random-looking
// value" methods) and SimRNG, the implementation used everywhere in this
// project. Each SimRNG owns its own private generator, built from exactly
// one core.Seed (see rand.go) — so nothing outside the code that was
// actually handed that SimRNG can ever consume one of its calls.
//
// How it works:
// SimRNG wraps Go's math/rand.Rand, but with its own private, seeded
// instance rather than the shared global one (the right-hand side of the
// diagram above). Every method on RNG just forwards to that private
// instance, so the underlying PRNG algorithm does the real work — this
// package's job is only to make sure nothing in this project can reach
// randomness except through a value it was explicitly handed.
//
// Important:
//
//   - Do NOT call the top-level math/rand functions (rand.Intn, rand.Seed,
//     rand.Float64, ...) anywhere in this project's deterministic packages.
//     Those read from and write to the one shared global source — exactly
//     the "logger steals a call" problem described above.
//
// Determinism note:
// Two SimRNG values built from the same core.Seed, asked the same number
// of times in the same order, produce the exact same sequence of results —
// on any machine, every time. That's the one guarantee this whole package
// exists to provide.
//
// The one sentence to remember: this package isn't trying to generate
// random numbers in some new way — Go's math/rand already does that well.
// It's trying to control who's allowed to ask for one.
package randsim
