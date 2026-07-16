# 0003. Testing strategy and Definition of Done

- Status: Accepted
- Date: 2026-07-16

## Context

We are building in thin daily slices and want each one to be trustworthy before it lands on
`dev`. "It compiles" is not evidence it works. A security tool that edits users' manifests
and opens PRs has to be validated against real inputs, not mocks alone.

## Decision

Every slice meets this **Definition of Done** before merging to `dev`:

1. **Unit tests** for the logic, as Go table tests.
2. **Fixture-based end-to-end tests.** Sample projects live under `testdata/` (Go excludes
   that directory from builds). Each fixture is a real, minimal project (for example an npm
   app pinned to a known-vulnerable dependency). CLI output is asserted against **golden
   files**, regenerated with a `-update` flag.
3. **A real run.** The built binary is executed once against a fixture and observed, using
   the `verify` skill. Behaviour, not just green tests.
4. **`make check` and CI green.** `make check` runs fmt, vet, and `go test -race`. CI runs
   the same on every push and PR.
5. **Docs updated.** The ROADMAP checkbox is ticked and an ADR is added if a non-trivial
   decision was made.

## Consequences

- Fixtures double as living documentation of what the CLI handles.
- Golden tests make output changes visible and deliberate in review.
- Slightly more upfront work per slice, paid back by never shipping a broken `dev` and by
  catching regressions the moment they appear.
- Fixtures containing deliberately-vulnerable dependencies are data, never installed or
  executed by the test run; we parse them, we do not run them.
