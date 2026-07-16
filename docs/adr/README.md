# Architecture Decision Records

Short records of decisions that are annoying to reverse. One per file, numbered in order.
Keep them to roughly a page: the point is the decision and why, not an essay.

## Format

Each ADR has: **Status** (Proposed / Accepted / Superseded by NNNN), **Context** (the forces
at play), **Decision** (what we chose, in the active voice), **Consequences** (what this makes
easy and what it costs). Copy `template.md` to start a new one.

## Index

| # | Title | Status |
|---|-------|--------|
| [0001](0001-language-and-dependency-posture.md) | Language and dependency posture | Accepted |
| [0002](0002-findings-are-an-input.md) | Findings are an input, not a capability | Accepted |
| [0003](0003-testing-and-definition-of-done.md) | Testing strategy and Definition of Done | Accepted |

Write a new ADR when you make a call on: CLI verb/UX structure, how a fix is represented and
applied, how the CLI hands off to the Orchestrator/Connectors, adding a third-party
dependency, or the release/versioning model.
