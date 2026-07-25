---
title: verifi status
description: Show what is vulnerable in a project and the fix for each.
sidebar_position: 3
---

# verifi status

`verifi status` scans a project against the OSV advisory database and prints,
for each vulnerable package, whether your code imports it, the version to
upgrade to, what that upgrade clears, and the honest limits of what has been
checked. It is read-only: it never changes your project.

Run [`verifi update`](update.md) once first to download the advisory database.

## Usage

<!-- BEGIN SIGNATURE -->
```
verifi status <path>
```
<!-- END SIGNATURE -->

## Flags

<!-- BEGIN FLAGS -->
```
--json      Print findings as JSON
--db <dir>  Use a specific OSV database directory
--offline   Skip the registry check for published versions
```
<!-- END FLAGS -->

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
        - Code impact not checked: not yet verified whether your code uses a changed part of the package.
        - Behaviour not verified: run your tests after upgrading.

MEDIUM   minimist 1.2.0   direct
   used: not imported by your code, removing it may clear this
   CVE-2021-44906 (GHSA-xvch-5gv4-984h)   Prototype pollution in minimist
   Fix: upgrade to 1.2.6   confidence: advisory
        Clears GHSA-xvch-5gv4-984h (fixed in 1.2.6 per OSV), a patch bump from 1.2.0.
        - Code impact not checked: not yet verified whether your code uses a changed part of the package.
        - Behaviour not verified: run your tests after upgrading.

Next: impact (does the fix touch code you use) and behavioural checks.
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
