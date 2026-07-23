# 0005. Confidence ladder, one fix command, patches not forks

- Status: Accepted
- Date: 2026-07-24

## Context

The core emits fix candidates with reasons ([0004](0004-core-is-read-only-reasoning.md)). Three
questions came up as we designed the write side:

1. What does "confidence" mean. A candidate that passes the project's unit tests is not proven
   safe. Tests passing does not prove the features still work. Real confidence, the kind that
   justifies changing a user's code for them, needs evidence that the software still operates,
   which is behavioural, not structural.
2. How many fix commands there are. We were drifting toward `apply`, `patch`, and `remove` as
   separate verbs. The kind of change is a property of the candidate, not a choice the user
   should have to make.
3. Whether a backport means maintaining patched builds of every version. Keeping a patched fork
   of every minor release is a combinatorial maintenance trap.

## Decision

**Confidence is a statement of what has been proven, and auto-fix is gated on the top rung.**

- **Structural** (static, cheap): the version jump is small and no symbol the project uses
  changed. Proves nothing the project touches moved. Does not prove behaviour. Enough to
  *propose* a fix, not to apply it unattended.
- **Behavioural** (runs the project's own tests and the functions it actually calls, old version
  against new, and compares): proves the software still operates. The only rung that earns
  **auto-fix**.

Below behavioural, Verifi does not silently change a user's code. It opens a pull request the
user reviews. PR handoff is therefore not a separate feature, it is what a medium-confidence fix
looks like. Every candidate states its rung and what was and was not checked.

**One fix command.** There is one `verifi fix`. The candidate carries how it applies itself
(version bump, drop an unused package, or apply a backport patch), so the tool does the right
thing and the user does not pick a verb. Modifiers only: `--pr` (open a PR instead of writing
locally), `--dry-run` (show the change, write nothing). No separate `apply`, `patch`, or
`remove` commands.

**Backports are patches, not forks.** The default fix is moving to a safe version that already
exists on the registry, which stores nothing. A backport exists only for the stuck case where no
safe published version is reachable, and it is stored as a **diff keyed by (package, version,
advisory)**, applied to the installed version at fix time. We never maintain a patched build of
every version.

## Consequences

- The reasoning output gains a `confidence` rung (`structural` | `behavioural`) with the evidence
  that earned it, and only `behavioural` is eligible for unattended apply.
- Behavioural checking (running the user's own tests in a sandbox) is on the critical path to
  auto-fix, not an optional extra. It stays a later layer, but it is the layer that unlocks the
  product's point.
- The command surface stays small: `inspect`, `status`, `fix`. Fix behaviour is data-driven from
  the candidate.
- CodeFix stores a sparse set of patch diffs, not a fork matrix. See the master plan for how the
  fix tree (per package, universal) and the app graph (per project, private) stay separate.
