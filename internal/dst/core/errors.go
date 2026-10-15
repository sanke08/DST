package core

import "errors"

// ErrInvalidID is the error we return when a function needed a real ID
// (a NodeID, EventID, or Seed whose Valid() is true) but got the empty 0 value
// instead.
//
// Why one shared error instead of each package making its own?
// If every package invented its own "bad id" error, code in one package
// couldn't recognize the error from another. With one shared value, everyone
// can check for the same thing.
//
// How you'd use it — check the ID first, bail out if it's empty:
//
//	if !id.Valid() {
//	    return core.ErrInvalidID
//	}
//
// Important — how to CHECK for this error (this is the one Go trap here):
//
//   - Do NOT compare with ==. Use errors.Is instead:
//
//     if errors.Is(err, core.ErrInvalidID) { ... }   // correct
//     if err == core.ErrInvalidID          { ... }   // buggy, see below
//
// Why == breaks:
// The == check works fine at first. But it's normal in Go for one function to
// "wrap" an error to add context before passing it up, like this:
//
//	err = fmt.Errorf("loading node config: %w", core.ErrInvalidID)
//
// Now err is no longer ErrInvalidID itself — it's a NEW error with
// ErrInvalidID tucked inside it. So `err == core.ErrInvalidID` is suddenly
// false, even though the real problem ("the ID was invalid") hasn't changed at
// all. Your "if it's an invalid ID, do X" code silently stops working, and Go
// gives you no warning.
//
// errors.Is is built exactly for this: it peels back the wrapping layer by
// layer and checks each one, so it still finds ErrInvalidID inside. That's why
// adding context to an error never breaks a caller who used errors.Is.
var ErrInvalidID = errors.New("core: invalid id")
