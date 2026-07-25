---
title: Fixing vulnerabilities
description: From a finding to a fix.
sidebar_position: 2
---

# Fixing vulnerabilities

`verifi status` does not just tell you a package is vulnerable, it tells you the
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

`verifi fix` applies the recommendation for you. It previews by default and
writes nothing:

```sh
verifi fix path/to/project
```

Add `--apply` to make the change; it runs the package manager, which updates both
your manifest and lockfile:

```sh
verifi fix path/to/project --apply
```

Then re-run `verifi status` to confirm it is clear, and run your tests. See
[verifi fix](../commands/fix.md) for details.

Removing an unused vulnerable dependency, and opening the change as a pull
request you review, are on the way.
