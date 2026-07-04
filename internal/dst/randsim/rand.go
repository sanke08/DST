package randsim

import (
	"math/rand"

	"github.com/sanke08/DST/internal/dst/core"
)

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
// NewRNG, below). Nothing about it is shared with any other SimRNG or with
// math/rand's global state.
//
// What it does:
// Holds one private *rand.Rand and forwards every RNG method to it.
type SimRNG struct {
	r *rand.Rand
}

// NewRNG builds a SimRNG whose entire future sequence of random-looking
// numbers is fixed by the single seed you pass in.
//
// Why this exists:
// This is the one and only doorway between "a Seed value" (from core) and
// "something you can actually draw random numbers from." Every later package
// that wants randomness (fault injection in Phase 7, for instance) gets its
// SimRNG from here — never by reaching for math/rand directly — so that the
// whole run's randomness traces back to exactly one recorded seed. That
// single-doorway rule is what makes "re-run seed 42 and get the identical
// run" possible at all.
//
// Why convert the seed to int64 instead of leaving it a uint64?
// core.Seed is a uint64, but Go's rand.NewSource wants an int64. These two
// types cover different ranges — a very large uint64 doesn't fit in an int64
// and wraps around to a negative number when converted. That is completely
// fine here, and worth being unbothered by: the conversion is a fixed,
// pure function of the seed, so the SAME core.Seed always produces the SAME
// int64, and therefore the SAME number sequence. Determinism only requires
// that the mapping never changes between runs — not that it preserves the
// numeric value. Two different seeds could in principle collide to the same
// int64, but that's a seed-space concern, not a determinism one; a given
// seed is still perfectly reproducible.
//
// Example:
// Think of it like dialing the starting code into the combination-lock toy
// from core.Seed's comment: NewRNG(core.Seed(42)) hands you back a generator
// already wound to seed 42's starting position, ready to spit out 42's fixed
// sequence — 83, 17, 56, ... — the same order every single time you dial 42.
//
//	NewRNG(core.Seed(42)) ⇒ a *SimRNG that always yields seed 42's sequence.
//
// Determinism note:
// This constructor performs no I/O, reads no clock, and consults no global
// state — the returned SimRNG's behavior is a pure function of seed alone.
// That is the entire reason the caller is forced to supply a seed here
// rather than letting the package pick one: a package-chosen seed (from
// time.Now(), say) would make every run start from a different place and
// destroy reproducibility.
func NewRNG(seed core.Seed) *SimRNG {
	// int64(seed) may wrap a large uint64 to a negative value — deliberately
	// harmless, for the reason spelled out in the "Why convert" section above.
	return &SimRNG{r: rand.New(rand.NewSource(int64(seed)))}
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
