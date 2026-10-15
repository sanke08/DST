// Package randsim gives the rest of the project a FAKE, controlled source of
// random numbers instead of the real, uncontrolled one.
//
// First, clear up a common wrong guess:
// "Same seed" does NOT mean "same number every time you ask." A seed gives you
// the same SEQUENCE of numbers, in the same order. Think of a shuffled deck of
// cards: the same shuffle gives you the same order of cards, not the same card
// on every draw. So seed 42 might always give:
//
//	1st time you ask => 83
//	2nd time you ask => 17
//	3rd time you ask => 56
//
// Start over from seed 42 and you get 83, then 17, then 56 again — same order,
// every run.
//
// Why we can't just use Go's normal math/rand:
// Go's global rand.Intn / rand.Float64 all draw from ONE generator shared by
// the whole program. Because the number you get depends on your POSITION in
// the sequence, if some unrelated code also draws from it, it shifts everyone
// else's position. Here's the bug that causes:
//
//	Scheduler asks (1st) => 83   — good
//	Logger asks    (2nd) => 17   — an unrelated log line stole this number!
//	Scheduler asks (3rd) => 56   — scheduler expected 17, but got 56
//
// The seed never changed. Nobody touched the simulation. Just adding a log
// line elsewhere changed the result. A repeatable simulator cannot allow that.
//
// The fix — give the simulation its OWN private generator that nothing else
// can reach:
//
//	the broken shared way              the way this package does it
//
//	   Program                             Simulation
//	  /   |    \                               │
//	Sim Logger Metrics                         ▼
//	  \   |    /                             SimRNG   (private, nobody else has it)
//	      ▼                                    │
//	  Global RNG                               ▼
//	  (everyone shares)                    *rand.Rand
//
// On the left, three unrelated things pull from one shared pile, so any of them
// can steal a number from the others. On the right, the SimRNG belongs to one
// simulation only — nothing else in the program even has a reference to it, so
// nothing else can steal from it.
//
// What's inside:
//   - RNG    — a small interface: "give me a random number" (see rand.go)
//   - SimRNG — the real thing everyone uses; each one owns its own private
//     generator built from exactly one Seed
//
// Important:
//
//   - Do NOT call Go's global math/rand functions (rand.Intn, rand.Float64,
//     rand.Seed) anywhere in this project. Those all share the one global pile —
//     exactly the "logger steals a number" bug above. Always go through a SimRNG.
//
// (Determinism note: two SimRNGs built from the same Seed, asked the same
// questions in the same order, give the exact same answers — on any machine,
// every time. That is the whole promise of this package.)
//
// One sentence to remember: this package isn't inventing a new way to make
// random numbers — Go already does that fine. It's controlling WHO is allowed
// to ask for one.
package randsim
