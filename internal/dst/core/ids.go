package core

import "fmt"

// NodeID is just a number that names one machine in our simulation.
//
// Picture a small chat app running on 3 servers. We call them
// server 1, server 2, and server 3. NodeID is that number.
//
//	NodeID(1) => the first server
//	NodeID(2) => the second server
//	NodeID(3) => the third server
//
// Why not just use a plain number (uint64) then?
// Two reasons:
//
//  1. Giving it a name (NodeID) makes the code read like English.
//     "send this message from NodeID(1) to NodeID(2)" is clearer than
//     "send from 1 to 2" where you have to guess what the 1 and 2 mean.
//
//  2. It stops mix-up bugs. Later we also have EventID (see below), which
//     is ALSO just a number. If both were plain numbers, you could
//     accidentally pass an event number where a server number was expected
//     and nobody would notice. Because they are different named types, Go
//     refuses to build the code if you mix them up. Free bug protection.
//
// One special rule to remember:
// The number 0 is NOT a real server. We use 0 to mean "nobody filled this
// in yet / empty / not set."
//
//	NodeID(3) => a real server, server number 3
//	NodeID(0) => "no server chosen", NOT a real server
//
// Think of it like a locker at the gym. Locker #3 is a real locker you can
// open. "Locker #0" means you haven't picked a locker yet.
//
// Why bother reserving 0?
// In Go, if you make a struct and forget to fill in a NodeID field, Go
// quietly sets it to 0 for you. If 0 were a real server, that forgotten
// field would look like "oh, this belongs to server 0" and your program
// would keep running with wrong data — a silent bug that's really hard to
// find. By saying "0 is never real", we can add a Valid() check that
// catches the mistake immediately.
//
// So: after you receive a NodeID from somewhere else, call Valid() on it
// before trusting it. Go's compiler will NOT catch a 0 for you — that's on
// us to check.
//
// (Determinism note, for later: server numbers are handed out by us in a
// fixed order — 1, 2, 3... — never picked randomly or from the clock. That
// way, re-running the simulation always gives the same numbers.)
type NodeID uint64

// Valid returns true if n is a real server (any number except 0).
//
//	NodeID(0).Valid() => false  (empty / not set)
//	NodeID(1).Valid() => true   (a real server)
func (n NodeID) Valid() bool {
	return n != 0
}

// String is how a NodeID looks when printed in logs.
//
//	NodeID(3).String() => "node-3"
func (n NodeID) String() string {
	return fmt.Sprintf("node-%d", uint64(n))
}

// EventID is just a number that names one thing that is scheduled to happen
// in the simulation.
//
// Examples of an "event": "deliver this chat message", "fire this timer",
// "this server wakes up". Each one gets its own EventID number so we can
// point at it later.
//
// Real-world use: say a server sets a 5-second timer, then the reply arrives
// early so we want to cancel that timer. We need a way to say WHICH timer to
// cancel — that's what the EventID is for:
//
//	EventID(42) => the specific event we are talking about
//	EventID(0)  => "no event", NOT a real one (same 0 rule as NodeID above)
//
// It's a separate type from NodeID for the same reason as before: so you
// can't accidentally swap "which server" and "which event" — Go won't let
// you.
//
// (Determinism note: EventIDs are handed out by the scheduler with a simple
// counter — 1, 2, 3... — one per event, never random. Same list of events
// scheduled => same EventIDs every time.)
type EventID uint64

// Valid returns true if e is a real event (any number except 0).
//
//	EventID(0).Valid()  => false  (empty / not set)
//	EventID(42).Valid() => true   (a real event)
func (e EventID) Valid() bool {
	return e != 0
}

// String is how an EventID looks when printed in logs.
//
//	EventID(42).String() => "event-42"
func (e EventID) String() string {
	return fmt.Sprintf("event-%d", uint64(e))
}

// Seed is the one starting number that makes a whole simulation repeatable.
//
// A computer's "random" numbers aren't truly random — they come from a
// formula that starts from a seed number. Give it the SAME seed and you get
// the SAME sequence of numbers back, every single time.
//
// Think of a combination-lock toy: type in the same starting code and it
// spits out the exact same "random" sequence each time. The seed is that
// starting code.
//
//	Seed(42) => run the simulation and you get the exact same result again
//	Seed(0)  => "no seed chosen", NOT a real one (same 0 rule as above)
//
// Why this matters for us: this is the whole point of the project. A test
// fails? Write down its seed. Re-run with that same seed and the failure
// happens again in exactly the same way, so you can actually debug it.
//
// Heads up: unlike our IDs, most random-number formulas are perfectly happy
// to start from 0. THEY don't treat 0 as special — WE do, only to catch the
// "forgot to set it" mistake. Valid() is what enforces our rule.
//
// (Determinism note: the seed must come from something we chose on purpose —
// a command-line flag, a saved file — never from time.Now() or the system's
// built-in randomness. If the seed itself were random, nothing would be
// repeatable.)
type Seed uint64

// Valid returns true if s was actually set (any number except 0).
//
//	Seed(0).Valid() => false  (caller must pick a real seed on purpose)
func (s Seed) Valid() bool {
	return s != 0
}

// String is how a Seed looks when printed in logs.
//
//	Seed(42).String() => "seed-42"
func (s Seed) String() string {
	return fmt.Sprintf("seed-%d", uint64(s))
}
