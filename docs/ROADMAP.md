# Roadmap

How we work: thin vertical slices, one runnable and tested increment at a time (see the
`ship-slice` skill and the Definition of Done in [adr/0003](adr/0003-testing-and-definition-of-done.md)).
Verifi is the fix layer; finding is an input we consume ([adr/0002](adr/0002-findings-are-an-input.md)).

Checkboxes are the source of truth for progress. Tick one only when its slice meets the DoD
on `dev`.

## Done (walking skeleton)

- [x] `verifi` binary, `welcome` splash, `version`, help
- [x] CI (vet/build/test), GoReleaser release pipeline, `install.sh` with checksum verify
- [x] Repo, module path, and docs at `github.com/verifisecurity/verifi`

## Next two weeks (fix pipeline)

Each slice is end-to-end and shippable on its own.

- [ ] **1. Inventory.** `verifi inspect <path>` resolves an npm project's dependencies
      (direct + transitive) to a structured inventory (`--json`). Fixture: a small npm app.
- [ ] **2. Consume OSV.** Annotate the inventory with what OSV reports as vulnerable. Findings
      are attributed to OSV with ids and versions (evidence-bound).
- [ ] **3. Status.** `verifi status` prints a human-readable "what needs fixing" view.
- [ ] **4. Propose a fix.** Compute the nearest safe version for each affected package; show
      the proposed change. No writes yet.
- [ ] **5. Apply a fix.** `verifi fix` writes the version bump to the manifest and lockfile.
- [ ] **6. Emit a patch.** `verifi fix --patch` outputs the change as a diff for review.
- [ ] **7. Gate a build.** CI mode exits non-zero on unfixed policy violations.
- [ ] **8. Second ecosystem.** Bring PyPI through the same inventory -> OSV -> fix pipeline.
- [ ] **9. PR handoff.** Open (or stub) the handoff that turns an applied fix into a PR via
      Connectors.
- [ ] **10. Polish.** Config file, consistent `--json`, docs and examples.

## Later (not scheduled)

- Maven/Gradle and Go module support
- Reachability signal to prioritise fixes that matter
- Deeper Orchestrator handoff (incident-response workflows)

Add a dated note here when the plan changes, so the reasoning is not lost.
