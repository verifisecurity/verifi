---
title: Usage signal
description: Why an unused vulnerable dependency is the easiest fix.
sidebar_position: 2
---

# Usage signal

`verifi status` scans your project's own source and marks each vulnerable
package by how your code relates to it:

- **imported by your code**: your source imports it directly. This is where a
  fix matters most.
- **pulled in indirectly**: a dependency of a dependency. Your code does not
  import it; something you depend on does.
- **not imported by your code**: a direct dependency your own code never
  imports. A strong candidate to remove.

## Why it is useful

Two reasons. First, it prioritises: a vulnerable package your code actually
imports deserves attention before one that never touches your app. Second, it
finds free wins: a direct dependency you never import can often just be removed,
which clears the vulnerability with no upgrade and no compatibility risk. Most
scanners never point this out.

## What it is, and is not

This is import-level usage: does your code reference the package at all. It is
the cheap first half of reachability. It is not yet symbol-level reachability,
whether the specific vulnerable function is reachable through your call graph;
that is deeper analysis that raises the [confidence](confidence.md) rung, and it
is on the roadmap.

Detection is heuristic: dynamic imports and plugin loaders can hide a real use,
so "not imported" is a strong hint to review, not an automatic delete.
