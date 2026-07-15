# Contributing to Verifi CLI

Thanks for your interest in improving the Verifi CLI. This document covers how to get
changes accepted.

## Ground rules

- **Be respectful.** Assume good faith and keep discussion technical.
- **Security issues do not go here.** Never file a vulnerability as a public issue —
  follow [SECURITY.md](SECURITY.md) instead.
- **Keep it lean.** The CLI is intentionally small and focused. Prefer reusing mature,
  well-maintained libraries over hand-rolling, and keep new dependencies minimal.

## Workflow

1. **Open an issue first** for anything non-trivial, so we can agree on the approach
   before you invest time.
2. **Fork and branch** from `main`. Use a short, descriptive branch name.
3. **Make focused commits.** One logical change per commit; write clear messages in the
   imperative mood ("Add scan retry", not "added retries").
4. **Open a pull request** against `main`. Describe what changed and why, and link the
   issue. Keep PRs small and reviewable.
5. CI must be green and at least one maintainer must approve before merge.

## Commit sign-off / DCO

By submitting a contribution you certify that you wrote it (or have the right to submit
it) and agree to license it under the project's [Apache-2.0 license](LICENSE). We may
require a `Signed-off-by` line (`git commit -s`) as a Developer Certificate of Origin.

## Local development

The runtime and toolchain are being finalized (single-binary, Go/Rust candidate). Build
and test instructions will be documented here once the toolchain is locked. Until then,
open an issue if you'd like to help shape the foundation.

## License of contributions

All contributions are accepted under the terms of the [Apache License 2.0](LICENSE).
