package core

import "errors"

// ErrInvalidID is returned when a function required a real NodeID, EventID,
// or Seed (something whose Valid() is true) but was handed the zero value
// instead.
//
// Why this exists:
// Every ID type in this package (NodeID, EventID, Seed) shares one
// zero-means-invalid convention — see NodeID's doc comment in ids.go for the
// full explanation. Once a function validates one of these and finds it
// zero, it needs one shared, importable error to return, instead of every
// package downstream inventing its own "bad id" error that wouldn't compare
// equal to anyone else's.
//
// What it does:
// A single sentinel error value, shared by all three ID types rather than
// one sentinel each — a caller doesn't usually care whether it was a
// NodeID, an EventID, or a Seed that was invalid, just that whatever ID it
// tried to use wasn't real.
//
// Example:
// A function that receives a NodeID from outside (a config file, an
// incoming message) checks it before doing anything else:
//
//	if !id.Valid() { return core.ErrInvalidID }
//
// What goes wrong if you compare with == instead of errors.Is:
// Suppose a caller does exactly what feels natural at first:
//
//	if err == core.ErrInvalidID { ... }
//
// That works — right up until some other function wraps the error to add
// context, which is completely normal Go:
//
//	err = fmt.Errorf("loading node config: %w", core.ErrInvalidID)
//
// Now `err == core.ErrInvalidID` is false. Not because the underlying
// problem changed — it's still exactly "the ID was invalid" — but because
// `err` is now a different value: a wrapper holding ErrInvalidID inside it,
// not ErrInvalidID itself. The caller's "if it's an invalid ID, do X" logic
// silently stops firing, with no compiler error to catch the mistake.
// `errors.Is(err, core.ErrInvalidID)` avoids this because it doesn't just
// compare values — it unwraps `err` layer by layer, checking each one,
// specifically so wrapping an error for context never breaks a caller's
// ability to recognize it.
//
// Important:
//
//   - Do NOT compare an error to ErrInvalidID with ==. Use
//     errors.Is(err, core.ErrInvalidID) instead, for the reason above.
var ErrInvalidID = errors.New("core: invalid id")
