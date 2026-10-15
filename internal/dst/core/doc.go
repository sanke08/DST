// Package core holds the small, shared words (types and errors) that every
// other package in this project needs to use.
//
// Think of it like the glossary at the front of a rulebook. Before anyone can
// play the game, everybody has to agree on what the basic words mean. That's
// all core does — it defines the words, it doesn't play the game.
//
// What's inside:
//   - NodeID   — a number that names one machine (see ids.go)
//   - EventID  — a number that names one scheduled thing that will happen
//   - Seed     — the one starting number that makes a run repeatable
//   - ErrInvalidID — the error used when one of those IDs is missing (errors.go)
//
// Why put them all here instead of in each package?
// Imagine two different packages each made up their OWN "NodeID". They'd look
// the same but Go would treat them as two unrelated types — you couldn't pass
// one where the other was expected, and their errors wouldn't match. Keeping
// these words in ONE place means the whole project speaks the same language.
//
// One hard rule for this package:
//
// Important:
//
//   - Do NOT import any other package from this project into core. core sits at
//     the very bottom — everyone depends on it, and it depends on no one. If
//     core imported something that (even indirectly) imported core back, Go
//     would refuse to build the whole project ("import cycle"). So core only
//     ever uses Go's own standard library.
//   - Do NOT add a type here "just in case" a future phase needs it. Add a type
//     only when the phase you're building right now actually uses it.
//
// (Determinism note: core is just type definitions — no clock, no randomness,
// no ordering — so there is nothing here that could behave differently between
// two runs. But other packages build their repeatable behavior on top of these
// types, especially Seed, so once code depends on them their meaning must stay
// stable.)
package core
