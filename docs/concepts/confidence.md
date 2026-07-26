---
title: Confidence
description: What "checked" means on a fix, and what it does not.
sidebar_position: 1
---

# Confidence

Every fix Verifi recommends states what has actually been checked, and what has
not. The point is honesty: you are never told more certainty than the evidence
supports.

Today Verifi reports at the **advisory** level: the recommended version clears
the advisory, and the size of the version jump is known (patch, minor, or
major). That is what the advisory database and the registry can prove.

What advisory does not check:

- whether your own code uses a part of the package that changed between the two
  versions, and
- whether your application still behaves the same after the change.

So an advisory-level fix is one to review and test, not to apply blind. Every
recommendation lists these limits explicitly, and Verifi never claims more than
it has checked.
