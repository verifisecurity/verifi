---
title: Fixing vulnerabilities
description: From a finding to a fix.
sidebar_position: 2
---

# Fixing vulnerabilities

`verifi scan` does not just tell you a package is vulnerable, it tells you the
fix. For each vulnerable package it recommends a corrective action:

- **Upgrade** to the nearest safe version. The common case. Verifi picks the
  nearest fix that clears the advisories and that actually exists on the
  registry, not blindly the latest.
- **Remove** the dependency. If it is a direct dependency your code never
  imports, removing it clears the vulnerability at zero compatibility risk. The
  [usage signal](../concepts/usage-signal.md) is what identifies these.

Every recommendation states its [confidence](../concepts/confidence.md) and what
it has not checked, so you know whether to apply it directly or look closer
first.

## Applying a fix

Scanning writes a plan and marks nothing in it. You decide what to apply by
editing that file, then `verifi fix` runs what you marked:

```sh
verifi scan path/to/project     # writes the plan, tells you where it is
$EDITOR ~/.verifi/projects/<project>/candidates.json
verifi fix path/to/project      # runs only what you marked
```

Set `"apply": true` on the fixes you want. With nothing marked, `fix` writes
nothing and says so, which means running it by accident is harmless.

The plan is the preview, so there is no preview flag. Unlike terminal output it
is a real artifact: you can read it, diff it, and keep it, and `scan --out` can
write it into the repository if you would rather review it like any other change.

`verifi fix` handles both an upgrade and a removal, whichever the scan
recommends for each package. It runs the package manager rather than editing
your lockfile. Then re-run `verifi scan` to confirm it is clear, and run your
tests. See [verifi fix](../commands/fix.md) for details, including why a
transitive dependency can only be proposed for now.

Opening the change as a pull request you review is on the way.
