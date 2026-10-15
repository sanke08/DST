package randsim

import (
	"math/rand"

	"github.com/sanke08/DST/internal/dst/core"
)

// RNG is anything that can hand back a random-looking number.
//
// Why it's an interface (and not just the real type):
// Later code that needs "a random delay" or "a 1-in-10 chance of dropping a
// message" asks for an RNG. Because it's an interface, that code doesn't know
// or care what's behind it — in a real run it's a seeded SimRNG, but you could
// also give it a fake that always returns the same number to make something
// easy to check. The code using it stays the same either way.
//
// Think of RNG like a bag of numbered tickets you draw from:
//
//	rng.Intn(6)   => one of 0, 1, 2, 3, 4, 5
//	rng.Float64() => some fraction from 0.0 up to (but not including) 1.0
//
// It only has the two kinds of randomness the project needs so far. We'll add
// more methods only when a later phase actually needs them, not before.
type RNG interface {
	Intn(n int) int
	Float64() float64
}

// SimRNG is the real random-number generator used everywhere in this project.
//
// Each SimRNG holds its OWN private generator (the r field below), built from
// one seed. Nothing outside the code that was handed this SimRNG can reach that
// generator — which is the whole point (see this package's doc.go for why that
// matters).
type SimRNG struct {
	// r is this SimRNG's own private number generator. Private (lowercase)
	// on purpose: nobody outside should touch it, or they could change what
	// numbers come out and break repeatability.
	r *rand.Rand
}

// NewRNG builds a SimRNG whose entire future stream of numbers is decided by
// the one seed you pass in.
//
// This is the ONLY way to make a SimRNG. That's deliberate: every bit of
// randomness in the whole run traces back to one seed you chose on purpose, so
// "re-run with seed 42 and get the identical run" actually works.
//
// Think of it like dialing the starting code into a combination-lock toy:
//
//	NewRNG(core.Seed(42)) => a generator wound to seed 42's starting point,
//	                         ready to give 42's fixed sequence (83, 17, 56, ...)
//
// One small thing about the int64(seed) below:
// core.Seed is a uint64, but Go's rand.NewSource wants an int64. A very large
// uint64 doesn't fit in an int64 and "wraps around" to a negative number when
// converted. That is totally fine here. All we need is that the SAME seed
// always turns into the SAME int64 (it does — it's a fixed conversion), so the
// same seed always gives the same numbers. We don't need the number itself to
// stay the same, just the behavior.
//
// (Determinism note: this does no I/O, reads no clock, uses nothing global —
// the result depends only on the seed. That's exactly why the caller must hand
// us a seed instead of us picking one from, say, the current time.)
func NewRNG(seed core.Seed) *SimRNG {
	// int64(seed) may wrap a large value to a negative one — harmless, as
	// explained just above.
	return &SimRNG{r: rand.New(rand.NewSource(int64(seed)))}
}

// Intn returns a random whole number from 0 up to (but not including) n.
//
//	rng.Intn(6) => some number in 0, 1, 2, 3, 4, 5
func (s *SimRNG) Intn(n int) int {
	return s.r.Intn(n)
}

// Float64 returns a random fraction from 0.0 up to (but not including) 1.0.
//
//	rng.Float64() => e.g. 0.37
func (s *SimRNG) Float64() float64 {
	return s.r.Float64()
}
