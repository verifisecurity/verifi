# Roadmap

How we work: thin vertical slices, one runnable and tested increment at a time (see the
`ship-slice` skill and the Definition of Done in [adr/0003](adr/0003-testing-and-definition-of-done.md)).
Verifi is the fix layer; finding is an input we consume ([adr/0002](adr/0002-findings-are-an-input.md)).
The core is read-only reasoning: we tell you which fix is safe and why, before anything is
written ([adr/0004](adr/0004-core-is-read-only-reasoning.md)). The pipeline and its schemas are
documented in [pipeline.md](pipeline.md).

Checkboxes are the source of truth for progress. Tick one only when its slice meets the DoD
on `dev`.

## Done (walking skeleton)

- [x] `verifi` binary, `welcome` splash, `version`, help
- [x] CI (vet/build/test), GoReleaser release pipeline, `install.sh` with checksum verify
- [x] Repo, module path, and docs at `github.com/verifisecurity/verifi`

## Next: the read-only core (npm first)

Each slice is end-to-end and shippable on its own. Nothing in this section writes to the
user's repo.

- [x] **1. Inventory.** `verifi inspect <path>` resolves an npm project's dependencies
      (direct + transitive) to a structured inventory, emitted as a CycloneDX SBOM and `--json`.
      Fixture: a small npm app.
- [x] **2. Match.** Annotate the inventory with what OSV reports as vulnerable, attributed to
      OSV with advisory ids and affected versions (evidence-bound). Offline: a local OSV database
      directory, matched with a stdlib semver comparator. Fixture: a vulnerable npm app.
- [x] **3. Status.** `verifi status <path> --db <dir>` prints a human-readable "what needs
      fixing" view, grouped by package and ranked by severity, plus `--json`.
- [ ] **4. Candidates.** For each vulnerable package, compute the fix options (nearest safe
      version, and how much of the risk each one clears) and show them. No writes.
- [ ] **5. Reasoning.** For each candidate, say why it will or will not work from what we can
      know cheaply: advisory-recorded breaking changes and how far the version jump is, with
      honest limits ("behaviour not verified"). The differentiator in its first, light form
      ([adr/0004](adr/0004-core-is-read-only-reasoning.md)).
- [ ] **6. Usage signal.** Narrow the list to packages the project actually imports, so the
      ones that never touch your code drop out. First cut: npm import detection.

## Later: deeper reasoning (each needs its own ADR)

- Code-level impact: which changed parts of a package your code actually uses, so a candidate can
  name what specifically would break. Per ecosystem, needs analysis tooling.
- Behavioural check: run the project's tests, or the specific functions it uses, against the old
  and new versions and compare. Needs a sandbox to run code safely.
- Second ecosystem: bring PyPI through the same inventory-to-reasoning pipeline, then Maven/Gradle
  and Go modules.

## Later: features on top (the write side)

One command, `verifi fix`, gated on confidence ([adr/0005](adr/0005-confidence-ladder-and-one-fix-command.md)).
The candidate carries how it applies itself, so there is no separate apply/patch/remove verb.

- `verifi fix`: apply the change the candidate describes (version bump, drop an unused package,
  or apply a backport patch). Behavioural-confidence candidates can apply unattended; everything
  below that opens a PR the user reviews (`--pr`), and `--dry-run` writes nothing.
- Behavioural check: run the project's own tests and used functions against old and new versions
  in a sandbox, and compare. This is what earns unattended apply, so it is on the critical path,
  not a nice-to-have.
- Gate a build: CI mode exits non-zero on an unfixed policy violation.
- Polish: config file, consistent `--json`, docs and examples.

Add a dated note here when the plan changes, so the reasoning is not lost.
