---
title: Verifi CLI docs
description: Fix vulnerable dependencies without breaking your app.
sidebar_position: 0
---

# Verifi CLI

Verifi is the fix layer for your software supply chain. Every scanner hands you a
list of vulnerable dependencies; Verifi looks at how your project actually uses
each one and tells you the version to move to, what it clears, and what it has
not checked yet, so you get a short, honest list of fixes instead of a backlog
you are afraid to touch.

## Start here

- [Getting started](guide/getting-started.md): install, download the database, run your first scan.
- [Fixing vulnerabilities](guide/fixing.md): from a finding to a fix.

## Concepts

- [Confidence](concepts/confidence.md): what "checked" actually means, and what does not.
- [Usage signal](concepts/usage-signal.md): why an unused vulnerable dependency is the easiest fix.

## Command reference

- [inspect](commands/inspect.md): resolve the dependency tree, export an SBOM.
- [update](commands/update.md): download the advisory database.
- [status](commands/status.md): show what is vulnerable and the fix.
- [fix](commands/fix.md): apply the recommended fix, or preview it.
- [version](commands/version.md): print the version.

The examples in the command pages are captured from real runs against the test
fixtures, so they show exactly what the tool produces.
