package randsim

import "math/rand"

// RNG is anything that can hand back a random-looking number.
//
// Why this exists:
// Later packages that need "a random-looking delay" or "a 1-in-10 chance of
// dropping this message" (fault injection, in Phase 7) shouldn't call
// math/rand's global functions directly — see this package's doc comment
// for why. Taking an RNG as a constructor argument instead means those
// packages don't know or care whether they're talking to a real seeded
// PRNG or, in a test, something that always returns a fixed value.
//
// What it does:
// A small interface with the two shapes of randomness this project needs
// so far: a random integer in a range (Intn) and a random fraction between
// 0 and 1 (Float64). More methods get added only when a later phase
// actually needs them — not speculatively now.
//
// Example:
// Think of RNG like a bag of numbered tickets you draw from. Intn(6) is
// like drawing from a bag with tickets 0 through 5 in it; Float64() is like
// drawing a ticket marked somewhere between 0.0 and 1.0.
//
//	rng.Intn(6)    ⇒ one of 0, 1, 2, 3, 4, 5.
//	rng.Float64()  ⇒ some fraction, 0.0 up to (but never quite reaching) 1.0.
type RNG interface {
	Intn(n int) int
	Float64() float64
}

// SimRNG is the seeded random number generator used everywhere in this
// project instead of math/rand's global functions.
//
// Why this exists:
// This is the concrete type that makes RNG deterministic: a SimRNG wraps
// its own private *rand.Rand, built from exactly one core.Seed (see
// NewRNG, added in the next step of this phase). Nothing about it is
// shared with any other SimRNG or with math/rand's global state.
//
// What it does:
// Holds one private *rand.Rand and forwards every RNG method to it.
type SimRNG struct {
	r *rand.Rand
}

// Intn returns a random integer in the range [0, n).
//
// Example: rng.Intn(6) ⇒ some integer from 0 up to (not including) 6.
func (s *SimRNG) Intn(n int) int {
	return s.r.Intn(n)
}

// Float64 returns a random fraction in the range [0.0, 1.0).
//
// Example: rng.Float64() ⇒ some number from 0.0 up to (not including) 1.0.
func (s *SimRNG) Float64() float64 {
	return s.r.Float64()
}
