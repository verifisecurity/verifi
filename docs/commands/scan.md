---
title: verifi scan
description: Report what is vulnerable in a project and the fix for each.
sidebar_position: 1
---

# verifi scan

`verifi scan` resolves a project's full dependency tree, matches it against the
OSV advisory database, and reports for each vulnerable package whether your code
imports it, the version to move to, what that clears, and the honest limits of
what has been checked. It is read-only: it never changes your project.

It is the one read-only command. The tree views describe what you depend on, the
default view reports what is wrong with it, and `--download` fetches the
advisory database it matches against.

## Usage

<!-- BEGIN SIGNATURE -->
```
verifi scan <path>
```
<!-- END SIGNATURE -->

## Flags

<!-- BEGIN FLAGS -->
```
--json        Print the selected view as JSON
--inventory   Report the dependency tree instead of the findings
--sbom        Print the dependency tree as a CycloneDX SBOM
--download    Fetch the OSV database into the local cache first
--db <dir>    Use a specific OSV database directory
--offline     Skip the registry check for published versions
--out <file>  Write the plan here instead of the default location
```
<!-- END FLAGS -->

`--json` is a modifier on whichever view you selected, not a view of its own.

## The plan

Every scan writes a plan: every fix it found, with the evidence behind it and
the gate's verdict, each entry marked `"apply": false`.

```json
{
  "apply": false,
  "package": "lodash",
  "current": "4.17.11",
  "action": "upgrade",
  "target": "4.17.21",
  "command": ["npm", "install", "lodash@4.17.21"],
  "distance": "patch",
  "clears": ["GHSA-35jh-r3h4-6jhm"],
  "used": "imported by your code",
  "gate": { "rung": "advisory", "authorization": "confirm" }
}
```

Open it, set `"apply": true` on the fixes you want, and run
[`verifi fix`](fix.md). Nothing is applied until you do: a scan never marks
anything, so a fix straight after a scan is a no-op.

The plan lives under `~/.verifi/projects/`, keyed to the project, not inside the
project itself, so scanning a repository never changes it. `scan` prints the
path each time. Use `--out <file>` to write it somewhere else, for instance into
the repository so a team can review it like any other change.

The plan also records what it was computed from: the advisory source, when that
data was fetched, and the version of verifi that wrote it.

## The first scan

Matching happens against a local copy of the OSV advisory database, stored in
`~/.verifi/osv` and shared by every project on the machine. If it is not there
yet, fetch it:

```sh
verifi scan path/to/project --download
```

That download is large and takes a few minutes. It is a one-off: later scans
reuse the cache and run offline. `verifi scan` tells you how old the cache is
and warns once it is stale, so refresh it with `--download` occasionally.

To point at a database you manage yourself, pass `--db <dir>` instead. Databases
supplied that way are used as given and never age-checked.

## Example

<!-- BEGIN EXAMPLE -->
```
vuln-app@1.0.0 (npm)
3 packages scanned, 2 vulnerable.

HIGH     lodash 4.17.11   direct
   used: imported by your code
   CVE-2021-23337 (GHSA-35jh-r3h4-6jhm)   Command injection in lodash
   Fix: upgrade to 4.17.21   confidence: advisory
        Clears GHSA-35jh-r3h4-6jhm (fixed in 4.17.21 per OSV), a patch bump from 4.17.11.
        - Code impact not checked: whether your code uses a part of the package that changed is not checked.
        - Not tested: run your tests after upgrading.

MEDIUM   minimist 1.2.0   direct
   used: not imported by your code, removing it may clear this
   CVE-2021-44906 (GHSA-xvch-5gv4-984h)   Prototype pollution in minimist
   Fix: remove minimist   confidence: advisory
        Your code does not import it; removing it clears GHSA-xvch-5gv4-984h at no compatibility risk.
        - Usage is heuristic: a dynamic import or a config-referenced loader can hide a real use, so confirm it is unused.

Plan written to ~/.verifi/projects/<project>/candidates.json
Mark the fixes you want with "apply": true, then run: verifi fix testdata/npm/vuln
```
<!-- END EXAMPLE -->

## Reading the output

For each vulnerable package you get:

- its worst severity, and whether it is a direct or transitive dependency,
- **used:** whether your code imports it, imported directly, pulled in
  indirectly, or a direct dependency you never import (a remove candidate),
- the advisories it is affected by, each with its id,
- **Fix:** the version to upgrade to and the confidence rung, then what that
  clears and what has not been checked.

Confidence is explained in [concepts: confidence](../concepts/confidence.md),
and the usage line in [concepts: usage signal](../concepts/usage-signal.md).

## Describing the tree

`--inventory` reports what you depend on rather than what is wrong with it, and
`--sbom` renders the same tree as a CycloneDX 1.5 SBOM. Both read only the
lockfile, so neither needs an advisory database.

```sh
verifi scan path/to/project --inventory
verifi scan path/to/project --inventory --json
verifi scan path/to/project --sbom > sbom.json
```
