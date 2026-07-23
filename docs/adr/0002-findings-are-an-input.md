# 0002. Findings are an input, not a capability

- Status: Accepted
- Date: 2026-07-16

## Context

Detecting known-vulnerable dependencies is commoditised. Public feeds (OSV, ecosystem
advisories) and existing tools (Dependabot, OSV-Scanner, Trivy, GitHub) already do it, for
free. Building another scanner would spend our effort where there is no differentiation and
no willingness to pay. The unmet pain is what happens *after* detection: teams accumulate a
backlog of findings they cannot action. Verifi's value is the fix and the orchestration of
the response.

## Decision

We will treat **finding as an input the CLI consumes, not a capability Verifi sells.** The
CLI resolves a project's dependencies and matches them against **existing sources** (OSV
first) to learn what needs fixing. It does not build, brand, or market a detection engine.
Its job starts where detection stops: propose a fix, apply it, gate the build, and hand the
response off to the Orchestrator.

Novel-malware detection (the Scanner) is a separate platform concern and out of scope for
this repo.

## Consequences

- Copy and UX lead with fix, remediation, and response. "Find/scan" appears only as the
  precondition it is, ideally attributed to the user's existing tools and feeds.
- We integrate with, rather than compete with, whatever already flags risk in a customer's
  stack.
- We depend on the quality and coverage of upstream feeds; where they are wrong or missing,
  we surface that rather than papering over it.
- Roadmap effort concentrates on the fix pipeline (resolve, propose, apply, PR, gate), not on
  detection breadth.

## Amendment (2026-07-24)

The CLI now produces findings itself: `verifi status` resolves the tree and matches it against a
local OSV database (slice A2). This is consistent with the decision, not a reversal. Matching a
resolved tree against OSV is consuming an existing source, which is exactly what "finding is an
input" means; we still built no detection engine and no scanner.

Two clarifications the original wording left open:

- Our own OSV match is the default and only required source. It covers the common case, so the
  fix pipeline never depends on the user running another tool first.
- Ingesting third-party findings (a scanner's SARIF, GitHub or Snyk output) remains an optional
  additional source for teams that already scan, and is deferred. It is additive, not a
  precondition.

The moat is unchanged: detection is commodity, the fix is the value.
