# 0001. Language and dependency posture

- Status: Accepted
- Date: 2026-07-16

## Context

The CLI ships as a product to end users and runs inside their build pipelines. It is the
public, open-source surface of a security company, so its own supply chain is on display: a
dependency we pull in is one our users inherit, and one a reviewer can point at. We want a
single self-contained binary, cross-platform, with fast cold starts and no runtime to
install.

## Decision

We will build the CLI in **Go**, and keep it **stdlib-first with zero third-party
dependencies** by default. A new dependency requires its own ADR justifying why the stdlib
is insufficient and why that specific library (maturity, maintenance, licence, transitive
weight) is worth adding.

We compile static binaries (CGO disabled) for linux, darwin, and windows on amd64 and arm64,
released via GoReleaser with a published checksums file.

## Consequences

- A security tool with a near-empty dependency tree is a feature we can point to, and it
  keeps our own attack surface minimal.
- Reproducible, dependency-free builds; trivial install (`curl | sh` to a single binary).
- We write more ourselves (for example the welcome splash is hand-rolled rather than using a
  TUI framework). When a real TUI or a hard problem (SBOM parsing, semver resolution) makes
  the stdlib genuinely costly, we revisit via a new ADR rather than reaching for a library by
  reflex.
- Go's cross-compilation and single-binary output match the distribution goal directly.
