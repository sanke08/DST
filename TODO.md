# Deterministic Simulation Testing (DST) Framework — Build Plan

Build rules (dependency ordering, determinism constraints, when to check a box) live in
[LEARNING.md](LEARNING.md) so this file stays a bare checklist.

---

- [x] **Phase 0 — Concurrency model decision**
  - DECISION: Single-threaded/event-driven SUT. A SUT implements callback-style
    hooks (`OnMessage`, `OnTimer`, `OnFault`) invoked one at a time by the
    scheduler. No SUT-owned goroutines.

- [x] **Phase 1 — Foundation** (`internal/dst/core`)
  - [x] Shared types/errors (`EventID`, `NodeID`, `Seed`, sentinel errors)

- [x] **Phase 2 — Deterministic primitives** (`internal/dst/timesim`, `internal/dst/randsim`)
  - [x] `Clock` interface + simulated implementation
  - [x] `RNG` interface wrapping a seeded PRNG

- [ ] **Phase 3 — Scheduler / event core** (`internal/dst/scheduler`)
  - [ ] Event queue + deterministic tie-breaking
  - [ ] `Step()` / `Run(until)`
  - [ ] Snapshot/restore

- [ ] **Phase 4 — Simulated transport & storage** (`internal/dst/transportsim`, `internal/dst/storagesim`)
  - [ ] Message send → scheduled delivery (delay/reorder/drop)
  - [ ] Disk write → scheduled completion (latency/corruption)

- [ ] **Phase 5 — Trace recording** (`internal/dst/trace`)
  - [ ] Trace format
  - [ ] Recorder hook into scheduler

- [ ] **Phase 6 — Replay** (`internal/dst/replay`)
  - [ ] Replay-from-seed
  - [ ] Replay-from-trace

- [ ] **Phase 7 — Fault injection** (`internal/dst/faults`)
  - [ ] Fault types (drop, delay, corrupt, crash-node)
  - [ ] Injector wired through transport/storage

- [ ] **Phase 8 — Invariants** (`internal/dst/invariants`)
  - [ ] Invariant interface
  - [ ] Checker (runs per-step or end-of-run)

- [ ] **Phase 9 — SUT harness** (`internal/dst/harness`)
  - [ ] Harness interface (per Phase 0 decision)
  - [ ] Runner composing scheduler + transport + storage + faults + invariants + trace

- [ ] **Phase 10 — Example SUT** (`examples/<name>`)
  - [ ] Pick and implement one small example
  - [ ] 2–3 invariants

- [ ] **Phase 11 — Reporting** (`internal/dst/report`)
  - [ ] Human-readable failure report

- [ ] **Phase 12 — CLI wiring** (`cmd/simulator`)
  - [ ] Flags: seed, fault profile, SUT selection, output format
  - [ ] `main.go` thin composition only

- [ ] **Phase 13 — Documentation**
  - [ ] README (incl. what this framework is NOT good for)
  - [ ] Per-package `doc.go`
  - [ ] "Add a new SUT" guide
  - [ ] "Reproduce a failure from a seed" guide
