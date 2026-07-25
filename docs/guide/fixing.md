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

Today you apply the recommendation yourself. For an upgrade:

```sh
npm install lodash@4.18.0
```

For an unused dependency:

```sh
npm uninstall left-pad
```

Then re-run `verifi status` to confirm it is clear.

## Coming soon: `verifi fix`

`verifi fix` will apply the recommended change for you, or open it as a pull
request you review, so you do not have to run the package manager by hand. It
previews by default and only writes when you opt in. See the roadmap on the
[repository](https://github.com/verifisecurity/verifi).
