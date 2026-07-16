# verifi (CLI)

The open-source **fix layer** for the software supply chain. Verifi takes what is already
flagged (by feeds and the tools you run) and turns it into a fix: a PR, or a gated build.
Finding is an input it consumes, not a capability it sells. This repo is the developer
surface, the `verifi` CLI. Pre-release, built in the open.

- Language: **Go**, stdlib-first, zero third-party deps unless an ADR justifies one (see
  [docs/adr/0001](docs/adr/0001-language-and-dependency-posture.md)).
- Canonical repo: `github.com/verifisecurity/verifi`. Binary: `verifi`.
- This file governs how we build this repo. The roadmap is [docs/ROADMAP.md](docs/ROADMAP.md);
  decisions are recorded in [docs/adr/](docs/adr/).

## How we build (binding on every session)

**Cadence: thin vertical slices.** Each change is one end-to-end increment that runs and is
tested, extending the working binary. Not horizontal layers, not aimless micro-commits. The
current plan is [docs/ROADMAP.md](docs/ROADMAP.md); the repeatable loop is the `ship-slice`
skill.

**Definition of Done** (nothing merges without all of these):
1. Unit tests for the logic (Go table tests).
2. A fixture-based end-to-end test: a sample project under `testdata/`, output asserted
   against golden files (`-update` to regenerate).
3. The real binary was run once against a fixture (use the `verify` skill).
4. `make check` is green (fmt, vet, test) and CI is green.
5. Docs updated: ROADMAP checkbox ticked; an ADR added if a non-trivial decision was made.

**Branching.** `dev` is the working branch (commit slices here). `main` is the release
branch (PR from `dev`, tag to release). Never push a red `dev`.

**Conventions.**
- No em dashes in any prose (docs, help text, comments). Use commas or periods.
- Every finding/fix the CLI emits cites concrete evidence (file, package, version, source),
  per the platform "evidence-bound" tenet. No unsourced claims.
- Decisions that are annoying to reverse get a short ADR in `docs/adr/` before or as we build.
