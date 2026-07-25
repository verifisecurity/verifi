---
title: Confidence
description: What "checked" means on a fix, and what it does not.
sidebar_position: 1
---

# Confidence

Every fix Verifi recommends carries a confidence rung that states what has
actually been proven. The point is honesty: you are never told more certainty
than we have.

## The rungs

- **advisory** (today's baseline): the advisory is cleared by this version, and
  we know how far the version jump is. Nothing about your own code or its
  behaviour has been checked. This is what `verifi status` reports now.
- **structural**: on top of the above, no symbol your code uses changed between
  the versions. It proves nothing you touch moved; it still does not prove
  behaviour. Needs code-level impact analysis, which is per ecosystem.
- **behavioural**: the project's own tests and the functions it actually uses
  pass on the new version. This proves the software still operates.

## Why it matters

Passing unit tests is not proof a fix is safe, and "upgrade to latest" is not
advice. The rung tells you how far to trust a recommendation. Only the
behavioural rung is strong enough to justify applying a change unattended;
below it, a fix is something to review, which is why automated fixes open a pull
request rather than changing your code silently.

Verifi deepens from advisory toward behavioural over time. It never claims a
higher rung than it has earned.
