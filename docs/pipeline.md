# The fix reasoning pipeline

How `verifi` turns a project into a short list of fix candidates, each with the reason it will
or will not work. This is the read-only core ([adr/0004](adr/0004-core-is-read-only-reasoning.md)):
it reads and reasons, it does not write to your repo or build your project. Every step is
evidence-bound: what it outputs cites its source (file, package, version, advisory id).

We build this npm-first, static-first. Behavioural checks (running the project) and the write
side (apply, patch, PR, gate) are later layers, noted at the end.

## Phases

| # | Step | What it does | Consumes | Emits |
|---|------|--------------|----------|-------|
| 1 | **Inventory** | Resolve the dependency tree, direct and transitive | project path | `Inventory` (+ CycloneDX SBOM) |
| 2 | **Match** | Mark which packages are vulnerable, from OSV | `Inventory` | `Finding[]` |
| 3 | **Usage** | Keep the findings the project actually uses, drop the rest | `Inventory` + project source | `Finding[]` (narrowed, with usage) |
| 4 | **Candidates** | For each finding, the fix options and how much each clears | `Finding[]` | `Candidate[]` |
| 5 | **Impact** | For each candidate, what it changes that you use | `Candidate` + `Inventory` (+ dep metadata) | `Impact` |
| 6 | **Reasoning** | Per candidate: will it work, limitations, why or why not, options | `Candidate` + `Impact` | `Recommendation` |

Output is a `status` view for people and `--json` for machines. Nothing is written to the repo.

Cheap steps gate expensive ones. Steps 1, 2 and 4 are pure data. Step 3 (usage) and step 5
(impact) are where analysis depth grows over time and per ecosystem; each real increase in depth
gets its own ADR.

## Schemas (contracts)

The shapes passed between steps. Ecosystem-agnostic: identity is a `purl` (Package URL). These
will move to `verifi-core` once a second consumer needs them; for now they live with the CLI.

- **Inventory** `{ ecosystem, root, packages: [ { purl, name, version, direct, scope, deps:[purl] } ] }`
- **Finding** `{ purl, name, version, advisory_ids:[], severity, affected_ranges:[], fixed_versions:[], source:"OSV", evidence:[] }`
- **Usage** `{ purl, used: true|false|unknown, sites:[ {file, line} ], evidence:[] }`
- **Candidate** `{ finding_ref, action: upgrade|replace|remove|backport, target, clears:[advisory_id], residual:[advisory_id], distance }`
- **Impact** `{ candidate_ref, breaking_changes:[ {symbol, kind, source} ], affects_used: bool, confidence, evidence:[] }`
- **Recommendation** `{ finding_ref, best: candidate_ref, confidence: structural|behavioural, reason, limitations:[], options:[candidate_ref], evidence:[] }`

The `action: backport` candidate applies a stored patch diff, not a version move, and only when
no safe published version is reachable ([adr/0005](adr/0005-confidence-ladder-and-one-fix-command.md)).
The `confidence` rung states what was proven: `structural` (no used symbol changed, enough to
propose) or `behavioural` (ran the project's tests, enough to auto-apply). Only `behavioural`
earns an unattended fix.

Every record carries `evidence[]`. A step that cannot determine something says so in the record
(for example `used: unknown`), it never leaves it out.

## Ecosystems

The pipeline is the same for every ecosystem. What changes per ecosystem is where the inventory
comes from and how deep the usage and impact analysis can go.

| Ecosystem | purl | Inventory source | OSV | Tier |
|-----------|------|------------------|-----|------|
| npm | `npm` | `package-lock.json` (parse) | yes | 1 (now) |
| PyPI | `pypi` | `poetry.lock` / `Pipfile.lock` / requirements | yes | 1 |
| Maven | `maven` | `mvn dependency:tree` (no lockfile) | yes | 1 |
| Go | `golang` | `go.mod` + `go.sum` | yes | 1 |
| Cargo | `cargo` | `Cargo.lock` | yes | 1 |
| NuGet | `nuget` | `packages.lock.json` / `project.assets.json` | yes | 1 |
| RubyGems | `gem` | `Gemfile.lock` | yes | 1 |
| Composer | `composer` | `composer.lock` | yes | 2 |
| Hex | `hex` | `mix.lock` | yes | 2 |
| Pub | `pub` | `pubspec.lock` | yes | 2 |
| Swift | `swift` | `Package.resolved` | yes | 2 |
| Conan | `conan` | `conan.lock` | partial | 2 |

Most ecosystems ship a lockfile we can parse with the standard library, no toolchain needed and
no third-party code run. Maven has no lockfile, so it invokes `mvn dependency:tree`; that needs
the toolchain present and is its own slice with its own ADR when we get there.

## Depth today, and what static means

- **Static now.** Steps 5 and 6 reason from what we can know without running the project:
  which version clears which advisory, how far the version jump is, and what the advisory or
  package metadata records as a breaking change. In npm this is lighter, because npm has weak
  static analysis; a candidate says "no advisory-listed breaking change" rather than "this exact
  call breaks".
- **Deeper static, later.** Code-level impact (the exact changed symbol your code uses) lands per
  ecosystem as the analysis for it is built, each with its own ADR.
- **Behavioural, later.** Running the project's tests or its used functions against the old and
  new versions, in a sandbox, to confirm behaviour held. Language-agnostic, and the strongest
  signal, but it runs code so it is a later layer.

## Not in this pipeline (later, features on top)

Apply a fix, emit a patch, open a PR, gate a build. These write or act; the pipeline above only
reads and reasons. See [ROADMAP.md](ROADMAP.md).
