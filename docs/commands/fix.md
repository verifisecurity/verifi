---
title: verifi fix
description: Apply the recommended fix, or preview it.
sidebar_position: 4
---

# verifi fix

`verifi fix` applies the fix that [`verifi status`](status.md) recommends:
upgrade a vulnerable package to a safe version, or remove a direct dependency
your code never imports. It **previews by default and writes nothing**; you opt
in to the change with `--apply`, which runs the package manager for you.

Applying is deliberately explicit. At today's [advisory
confidence](../concepts/confidence.md), Verifi does not change your code
silently; you ask for it, and then you run your tests.

## Usage

<!-- BEGIN SIGNATURE -->
```
verifi fix <path>
```
<!-- END SIGNATURE -->

## Flags

<!-- BEGIN FLAGS -->
```
--apply     Write the change via the package manager (default is preview)
--db <dir>  Use a specific OSV database directory
--offline   Skip the registry check for published versions
```
<!-- END FLAGS -->

## Example

A preview (the default), which writes nothing:

<!-- BEGIN EXAMPLE -->
```
Planned fixes (2). Nothing is written without --apply.

  upgrade lodash: 4.17.11 -> 4.17.21
      Clears GHSA-35jh-r3h4-6jhm (fixed in 4.17.21 per OSV), a patch bump from 4.17.11.
      $ npm install lodash@4.17.21

  remove minimist
      Your code does not import it; removing it clears GHSA-xvch-5gv4-984h at no compatibility risk.
      $ npm uninstall minimist

Apply with:  verifi fix <path> --apply
```
<!-- END EXAMPLE -->

## Applying

To make the changes, add `--apply`:

```sh
verifi fix path/to/project --apply
```

It runs the package manager for each fix (`npm install <pkg>@<version>` to
upgrade, `npm uninstall <pkg>` to remove), which updates both your manifest and
lockfile. Then re-run `verifi status` to confirm, and run your tests.

Opening the change as a pull request you review is on the way.
