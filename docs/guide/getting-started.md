---
title: Getting started
description: Install Verifi, download the database, run your first scan.
sidebar_position: 1
---

# Getting started

## Install

Verifi is a single self-contained binary with no runtime dependencies.

```sh
curl -fsSL https://raw.githubusercontent.com/verifisecurity/verifi/main/install.sh | sh
```

Or download a build from [Releases](https://github.com/verifisecurity/verifi/releases).

## Scan a project

Point `scan` at a project with a lockfile:

```sh
verifi scan path/to/project
```

You get, for each vulnerable package, whether your code imports it, the version
to upgrade to, what that clears, and the limits of what has been checked. Nothing
is written to your project; `scan` is read-only.

The first scan needs the OSV advisory database, which is stored in `~/.verifi/osv`
and shared by every project on the machine. If it is not there yet, `scan` tells
you so and `--download` fetches it:

```sh
verifi scan path/to/project --download
```

That download is large and takes a few minutes, once. Later scans reuse the cache
and run offline.

## What next

- Understand the output: [reading scan](../commands/scan.md#reading-the-output).
- Understand the confidence rung on each fix: [confidence](../concepts/confidence.md).
- Turn a finding into a change: [fixing vulnerabilities](fixing.md).
