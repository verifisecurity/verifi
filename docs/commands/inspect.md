---
title: verifi inspect
description: Resolve a project's dependencies into an inventory or SBOM.
sidebar_position: 2
---

# verifi inspect

`verifi inspect` resolves a project's full dependency tree, direct and
transitive, into a structured inventory. It reads the lockfile and does not
install or run anything. Use it to see what you actually depend on, or to export
a CycloneDX SBOM for another tool.

## Usage

<!-- BEGIN SIGNATURE -->
```
verifi inspect <path>
```
<!-- END SIGNATURE -->

## Flags

<!-- BEGIN FLAGS -->
```
--json  Print the inventory as JSON
--sbom  Print a CycloneDX SBOM
```
<!-- END FLAGS -->

## Example

<!-- BEGIN EXAMPLE -->
```
vuln-app@1.0.0 (npm)
3 packages: 3 direct, 0 transitive, 0 dev
  left-pad@1.3.0  [direct, prod]
  lodash@4.17.11  [direct, prod]
  minimist@1.2.0  [direct, prod]
```
<!-- END EXAMPLE -->

## Reading the output

Each line is a resolved package with its version and two tags: whether it is a
direct or transitive dependency, and whether it is a production or development
dependency. `--json` emits the same inventory as structured data; `--sbom`
emits a CycloneDX SBOM.
