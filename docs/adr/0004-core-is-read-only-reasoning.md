# 0004. The core is read-only fix reasoning

- Status: Accepted
- Date: 2026-07-21

## Context

Verifi's job starts where detection stops ([0002](0002-findings-are-an-input.md)). The obvious
next move is to apply fixes and open pull requests. But the hardest and most valuable question
for a user is not "can you open the PR", it is "which fix is safe for my project, and why".
Applying a fix writes to the user's repo and carries risk. Telling them which fix is safe carries
no risk, and it is the thing no scanner provides. We also build stdlib-first
([0001](0001-language-and-dependency-posture.md)), so deep code analysis arrives gradually and
per ecosystem, not on day one.

## Decision

We will make the CLI's core a **read-only reasoning step**. It resolves the project, matches
findings, and emits fix candidates, each with the reason it will or will not work, its
limitations, and the alternatives. This runs without writing to the user's repo and without
building their project.

Applying a fix, emitting a patch, opening a PR, gating a build, and behavioural verification are
**features layered on top of this core, not the core itself**. They come after.

The reasoning deepens over time and per ecosystem. It starts with what we can know cheaply and
stdlib-first: which versions resolve which advisories, how far the version jump is, and what the
advisory records as a breaking change. It grows toward code-level impact, meaning which changed
parts of a package the project actually uses, as the analysis for each ecosystem lands, each
addition with its own ADR.

## Consequences

- The first useful releases carry no risk to a user's repo. They read and reason, they do not
  write.
- Honesty is built in. A candidate states what has been checked and what has not, for example
  "no advisory-listed breaking change" versus "behaviour not verified, run your tests". We never
  imply more confidence than we have.
- Apply and PR handoff still ship, later, as a natural confidence upgrade on top of the reasoning.
- We are explicitly **not** leading with automated apply or PR, not building a scanner (0002),
  and not building our own build or test runner. When behaviour needs checking, we run the user's
  own build.
