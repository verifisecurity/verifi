---
title: verifi fix
description: Apply the fixes you marked in the plan.
sidebar_position: 2
---

# verifi fix

`verifi fix` runs the fixes you marked in the plan that
[`verifi scan`](scan.md) wrote: upgrade a vulnerable package to a safe version,
or remove a direct dependency your code never imports.

It analyses nothing of its own. Every command it runs was decided at scan time
and written into the plan, so what runs is exactly what you reviewed. There is
no preview flag because the plan is the preview.

## Usage

<!-- BEGIN SIGNATURE -->
```
verifi fix <path>
```
<!-- END SIGNATURE -->

## Flags

<!-- BEGIN FLAGS -->
```
--plan <file>  Read the plan from here instead of the default location
```
<!-- END FLAGS -->

## The loop

```sh
verifi scan path/to/project     # writes the plan, marks nothing
$EDITOR ~/.verifi/projects/<project>/candidates.json
verifi fix path/to/project      # runs only what you marked
```

Scan tells you where the plan is. Open it, set `"apply": true` on the fixes you
want, and run `fix`. With nothing marked, `fix` writes nothing and tells you so:

```
Nothing marked to apply in ~/.verifi/projects/<project>/candidates.json
It has 2 fix(es) available. Set "apply": true on the ones you want, then run this again.
```

## Two things have to agree

A fix runs only when **you marked it** and **the gate authorised it**. Either
alone does nothing.

Marking a fix says you want it. The gate is a separate question: whether verifi
can actually deliver it. Where it cannot, `fix` says so and moves on rather than
running a command that might not work:

```
Skipping lodash: marked, but verifi can only propose this one.
   Pulled in by another package, so installing it directly would add a top-level
   pin and might not clear the advisory. Forcing a transitive version needs an
   overrides entry; describing it only.
```

Each entry in the plan carries its `gate.authorization`, so you can see which
fixes are runnable before you mark anything.

## What it actually does

Verifi never edits your lockfile. It runs the package manager and lets npm write:

- **upgrade** runs `npm install <pkg>@<version>`
- **remove** runs `npm uninstall <pkg>`

npm then updates `package.json`, `package-lock.json`, and `node_modules`
together. Note that npm decides the range it saves: `npm install lodash@4.17.21`
pins `4.17.21` in the lockfile but writes `^4.17.21` into `package.json`, unless
you have `save-exact` set in your `.npmrc`.

Afterwards, re-run `verifi scan` to confirm the advisory is gone, and run your
tests. At today's [advisory confidence](../concepts/confidence.md), verifi has
checked that the target version clears the advisory and exists, not that it works
with your code.

## Limitations

**Transitive dependencies cannot be fixed yet.** A vulnerable package pulled in
by another package is always proposed, never applied. Installing it directly
would add a top-level entry for something you never depended on, and if its
parent's range excludes the new version npm keeps the vulnerable copy nested
underneath, so the advisory would survive a change that reported success. Doing
this properly needs an `overrides` entry, which is a different change and is not
built yet.

Opening the change as a pull request you review is on the way.
