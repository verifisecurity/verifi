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
