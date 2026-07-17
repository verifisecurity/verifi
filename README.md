# Verifi CLI

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

The command-line surface for [Verifi](https://verifisecurity.com), the open-source **fix
layer** for your software supply chain.

Findings come from anywhere: your scanners, public feeds, the registry, runtime. `verifi`
takes what is flagged, decides what matters by policy, and drives the fix in your project:
it opens the pull request and gates the build, right where you already work, in your
terminal and CI. It is not another scanner. Its job starts where detection stops.

Estate-wide response, blocking installs at the registry and coordinating fixes across
repositories, is the job of the wider Verifi platform the CLI plugs into.

> **Status: pre-release, built in the open.** The command surface is still being shaped and
> interfaces may change before the first tagged release. Watch
> [Releases](https://github.com/verifisecurity/verifi/releases) for the first stable cut.

## Install

_Coming with the first release._ Distribution will be a single self-contained binary with
no runtime dependencies.

## What it will do

- **Decide, don't just flag.** Resolve the full dependency tree, direct and transitive,
  cross-check known CVEs, end-of-life status, and known-malicious releases (from the public
  advisory databases or your own tools), and decide what matters by policy.
- **Fix it in your repo.** For anything with a known fix, bump or replace the package and
  open the pull request. The one step that actually moves the number.
- **Gate the rest.** No fix path yet? Exit non-zero on a policy violation, so nothing that
  breaks policy ships. One policy, two modes: fix what it can, gate the rest.
- **Hand off the rest.** Applied fixes become pull requests through Verifi Connectors.
  Registry-level blocking and cross-repo response belong to the Verifi platform, not the CLI.

## Usage

```
verifi fix <path>     # decide what matters, open fixes, gate the rest
verifi status         # show what needs fixing in this project
```

Full command reference lands with the first release.

## Contributing

Issues and pull requests are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md)
first, and note that contributions are accepted under the terms of the project license.

## Security

Found a vulnerability? **Do not open a public issue.** Follow the coordinated-disclosure
process in [SECURITY.md](SECURITY.md).

## License

Licensed under the [Apache License, Version 2.0](LICENSE).
Copyright © 2026 Verifi Security.
