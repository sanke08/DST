package core

import "fmt"

// NodeID identifies one participant ("node") in a simulated system — for
// example, one replica in a distributed key-value store.
//
// A common misconception, cleared up first:
// In most Go code, a zero value is just an ordinary, usable default — an
// empty string is a fine string, 0 is a fine int. NodeID does NOT work that
// way. NodeID(0) is reserved to mean "no node was set" — it is never a real
// node, even though nothing about it looks unusual to Go. Every real node
// must be assigned a NodeID of 1 or greater. This zero-means-invalid
// convention is shared by every ID type in this file (see EventID and Seed
// below), so it's worth getting comfortable with it here, once.
//
// Why this exists:
// Later packages (transport, harness) need to say "this message is from
// node 3, to node 7" without agreeing separately on what a node identifier
// even is. Wrapping a plain uint64 in a named type also stops you from
// accidentally passing an EventID where a NodeID was meant — the Go
// compiler rejects that at build time, even though both are uint64
// underneath.
//
// What goes wrong without this convention:
// In Go, a struct field of type NodeID that nobody explicitly set defaults
// to NodeID(0) automatically — there's no way to tell "someone deliberately
// chose node zero" apart from "this field was just never filled in." If
// NodeID(0) were treated as an ordinary, valid node, a forgotten field
// wouldn't cause a crash or an obvious error — it would silently behave as
// if "node 0" were a real participant in the simulation, and a bug (a
// config value never wired up) would look exactly like ordinary, correct
// data. Reserving 0 as "invalid" turns that silent bug into something
// Valid() (below) can actually catch.
//
// Example:
// Think of a NodeID like a locker number at a gym. Locker #3 points at
// one specific, real locker. Locker #0 doesn't exist — it just means "no
// locker picked yet."
//
//	NodeID(3) ⇒ a real node, node number 3.
//	NodeID(0) ⇒ means "no node," not a real one.
//
// Important:
//
//   - Do NOT assume a NodeID handed to you is valid just because it
//     type-checks. Go's compiler happily accepts NodeID(0) anywhere a
//     NodeID is expected — "zero means invalid" is a convention this
//     project agrees to, not something the type system enforces for you.
//     Call Valid() (defined right below) on any NodeID that came from
//     outside your own code before you trust it.
//
// Determinism note:
// NodeID values must be assigned by the caller (e.g. "node 1, 2, 3..." in
// the order the SUT is configured), never generated from time or
// randomness. As long as assignment order is fixed, NodeIDs are
// deterministic by construction.
//
// The one sentence to remember: in this package, zero doesn't mean "the
// default" — it means "nothing was set."
type NodeID uint64

// Valid reports whether n is a real node identifier (non-zero).
//
// Example: NodeID(0).Valid() ⇒ false; NodeID(1).Valid() ⇒ true.
func (n NodeID) Valid() bool {
	return n != 0
}

// String renders n for logs and error messages.
//
// Example: NodeID(3).String() ⇒ "node-3".
func (n NodeID) String() string {
	return fmt.Sprintf("node-%d", uint64(n))
}

// EventID identifies one scheduled event in the simulation — for example,
// "deliver this message" or "fire this timer."
//
// Why this exists:
// The scheduler (Phase 3) needs a way to name individual events so it can,
// for instance, cancel one specific pending timer, or so a trace (Phase 5)
// can record "event 42 fired at virtual time 100." Giving events their own
// ID type — rather than reusing NodeID or a bare uint64 — keeps "which
// node" and "which event" from being confused at a call site.
//
// What it does:
// Same zero-means-invalid convention as NodeID above, applied to scheduled
// events instead of nodes:
//
//	EventID(42) ⇒ a real, specific scheduled event.
//	EventID(0)  ⇒ means "no event," not a real one.
//
// Determinism note:
// EventIDs must be assigned by the scheduler in a fixed, reproducible order
// (e.g. a counter that increments once per scheduled event) — never from
// time or randomness — so that replaying the same sequence of scheduling
// calls produces the same EventIDs every time.
type EventID uint64

// Valid reports whether e is a real event identifier (non-zero).
//
// Example: EventID(0).Valid() ⇒ false; EventID(42).Valid() ⇒ true.
func (e EventID) Valid() bool {
	return e != 0
}

// String renders e for logs and error messages.
//
// Example: EventID(42).String() ⇒ "event-42".
func (e EventID) String() string {
	return fmt.Sprintf("event-%d", uint64(e))
}

// Seed is the starting value fed into a pseudo-random number generator
// (PRNG) to make an entire simulation run reproducible.
//
// Why this exists:
// Phase 2's RNG will be constructed from exactly one Seed value, and that's
// the master key that makes "run it again and get the exact same sequence
// of random-looking numbers" possible. Naming it here, in core, means Phase
// 2 (and every phase after it that needs to print or log "which seed
// produced this run") shares one definition instead of each package
// deciding for itself whether a seed is a uint64, an int64, or a string.
//
// What it does:
// Same zero-means-invalid convention as NodeID and EventID above — with one
// twist worth calling out: most PRNG algorithms treat 0 as a perfectly
// ordinary starting value. They have no idea this project wants to reserve
// it for "unset." Valid() (below) is what enforces that distinction.
//
// Example:
// Think of a Seed like the starting code you type into a combination-lock
// toy that spits out a "random" sequence of numbers. Type in the same
// starting code twice, and you get the exact same sequence back both
// times — it only looks random if you don't know the code.
//
//	Seed(42) ⇒ always the same sequence of "random" numbers, every run.
//	Seed(0)  ⇒ this project treats as "no seed chosen," not a real one.
//
// Determinism note:
// Seed is the root of determinism for the entire framework. Nothing
// upstream of a Seed value may be random — the seed itself must come from
// an explicit, recorded source (a CLI flag, a trace file), never from
// time.Now() or an unseeded system RNG.
type Seed uint64

// Valid reports whether s has been explicitly set (non-zero).
//
// Example: Seed(0).Valid() ⇒ false — this project requires callers to pick
// a real seed rather than silently defaulting to the zero value.
func (s Seed) Valid() bool {
	return s != 0
}

// String renders s for logs and error messages.
//
// Example: Seed(42).String() ⇒ "seed-42".
func (s Seed) String() string {
	return fmt.Sprintf("seed-%d", uint64(s))
}
