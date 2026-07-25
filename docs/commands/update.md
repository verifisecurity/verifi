---
title: verifi update
description: Download the OSV advisory database into the local cache.
sidebar_position: 1
---

# verifi update

`verifi update` downloads the OSV advisory database for an ecosystem into a
local cache at `~/.verifi/osv`. Run it once, from anywhere: it is machine-global,
not per-project. After it, [`verifi status`](status.md) matches offline, nothing
leaves your machine. Refresh it occasionally to pick up new advisories.

## Usage

<!-- BEGIN SIGNATURE -->
```
verifi update
```
<!-- END SIGNATURE -->

## Flags

<!-- BEGIN FLAGS -->
```
--ecosystem <name>  Ecosystem to download (default npm)
```
<!-- END FLAGS -->

## Output

It reports how many advisories it stored and where, for example:

```
Downloading the OSV database for npm ...
Stored 223721 advisories in /Users/you/.verifi/osv/npm
```

The advisory count changes as the database grows, so this example is
representative rather than captured.
