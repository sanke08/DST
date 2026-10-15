package scheduler

import "container/heap"

// eventQueue is the waiting line of events that haven't run yet, kept in the
// right order so the scheduler can always grab the one that should happen next.
//
// What "the right order" means:
//   - earliest FireTime first (time 100 runs before time 250)
//   - if two events share the EXACT same FireTime, the one scheduled first
//     (smaller seq) runs first — this is the tie-break the Event.seq field
//     exists for
//
// Why a heap instead of just sorting a list?
// We do two things over and over: add a new event, and take out the earliest
// one. If we kept a fully sorted list, every add would mean shuffling things to
// keep it sorted. A heap (Go's container/heap) keeps only enough order to answer
// "what's the smallest?" instantly and re-settle after each add/remove quickly.
// The important part for us is that it's DETERMINISTIC: given the same events
// in the same order, it always hands them back in the same order.
//
// Heads up on how Go's container/heap works: it does NOT give you a ready-made
// heap. It gives you the algorithm, and you supply the storage plus five small
// methods (Len, Less, Swap, Push, Pop) that tell it how to compare and move
// your items. Everything below is just those five methods over a plain slice.
//
// This type is lowercase (private) on purpose: it's an internal gear of the
// scheduler. Outside code schedules events through the Scheduler, never by
// poking the queue directly.
type eventQueue []*Event

// Len reports how many events are waiting. Required by container/heap.
func (q eventQueue) Len() int { return len(q) }

// Less is the whole ordering rule in one place: it answers "should event i run
// before event j?" The heap calls this constantly to keep the earliest event at
// the front.
//
//   - First compare FireTime: earlier time wins.
//   - Only when the times are exactly equal do we fall back to seq: smaller seq
//     (scheduled earlier) wins.
//
// This is exactly the two-level rule described on Event.seq, now actually
// enforced. Because both FireTime and seq are fixed values computed by the
// scheduler (never from the real clock or randomness), the answer here is the
// same on every run.
func (q eventQueue) Less(i, j int) bool {
	if q[i].FireTime.Equal(q[j].FireTime) {
		return q[i].seq < q[j].seq
	}
	return q[i].FireTime.Before(q[j].FireTime)
}

// Swap exchanges two events' positions. The heap calls this while it re-settles
// itself after an add or remove. Required by container/heap; nothing surprising.
func (q eventQueue) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

// Push adds one event to the underlying slice. NOTE: this is the low-level hook
// container/heap calls — you do NOT call it directly. To actually schedule an
// event you call heap.Push(&q, ev), which appends here and then re-settles the
// heap. The `any` parameter is Go's older interface style that container/heap
// still uses; we assert it back to *Event.
func (q *eventQueue) Push(x any) {
	*q = append(*q, x.(*Event))
}

// Pop removes and returns the LAST element of the slice. This too is a low-level
// hook: heap.Pop(&q) first moves the real smallest element to the end (using
// Swap/Less), then calls this to chop it off. So the value this returns really
// is the earliest event, even though the code here just grabs the tail.
//
// We also nil out the removed slot before shrinking, so the popped *Event isn't
// kept alive by the backing array — a small tidy-up that avoids a subtle leak.
func (q *eventQueue) Pop() any {
	old := *q
	n := len(old)
	ev := old[n-1]
	old[n-1] = nil
	*q = old[:n-1]
	return ev
}

// The blank assignment below is a compile-time check, not runnable code: it asks
// Go to confirm *eventQueue really satisfies heap.Interface (all five methods
// above). If we ever rename or mistype a method, the build fails right here with
// a clear message instead of somewhere confusing later.
var _ heap.Interface = (*eventQueue)(nil)
