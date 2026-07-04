// Package core holds the shared, dependency-free vocabulary of the DST
// framework: the ID types and sentinel errors that every other package in
// this project needs to refer to.
//
// Why this exists:
// Without a shared place for these types, every later package (scheduler,
// transport, storage, ...) would invent its own NodeID or its own "invalid
// ID" error, and then two packages' IDs wouldn't be comparable and their
// errors wouldn't match with errors.Is. Centralizing them here means the
// whole framework speaks the same vocabulary. It also keeps this package
// import-cycle-free forever: core sits at the very bottom of this project's
// dependency tree, so every other package depends on it, directly or
// indirectly, and it must never depend on any of them.
//
// What it does:
// Defines NodeID, EventID, and Seed — small value types that identify
// "which node," "which scheduled event," and "which random-number seed" —
// plus the sentinel errors returned when those values are used incorrectly
// (e.g. a zero-value ID passed where a real one is required).
//
// How it works:
// These are plain wrapper types around uint64, each with a Valid() method
// (zero value means "not set") and a String() method for logging. Later
// packages take these types as constructor/parameter types instead of bare
// uint64, so the compiler catches a NodeID accidentally passed where an
// EventID was expected.
//
// Important:
//
//   - Do NOT add an import to any other package in this project, even an
//     internal/dst one. The moment core imports something that imports core
//     back, the build breaks with an import cycle, and every downstream
//     package (all of them) inherits the mess.
//   - Do NOT add a type "just in case" a later phase might want it. Add a
//     type here only when the phase that's being built right now needs it.
//
// Determinism note:
// core defines types only, no logic that touches time, randomness, or
// ordering — so there's nothing here that can diverge between two runs by
// itself. But the types it defines (especially Seed) are the seam that
// later packages use to stay deterministic, so their meaning must stay
// stable once other packages start depending on them.
//
// A note on reading examples in this package: wherever a comment writes
// `X ⇒ Y`, read the ⇒ as "gives you" or "results in" — e.g. `NodeID(0) ⇒
// not a real node` just means "if you write NodeID(0), what you get is a
// value that isn't a real node."
//
// The one sentence to remember: this package doesn't do anything by
// itself — it just gives every other package one shared vocabulary
// (NodeID, EventID, Seed, and the error for when one of them is missing)
// so they don't each end up inventing a slightly different version of it.
package core
