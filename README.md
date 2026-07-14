# Verifi CLI

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

The open-source command-line interface for the [Verifi](https://verifisecurity.com)
supply-chain security platform.

`verifi` is a thin, cross-platform client. It carries no detection logic of its own —
it wraps the Verifi backend (detection → triage → remediation) over that backend's
published API, so the CLI stays small and auditable while the analysis runs server-side.
Use it to scan dependencies, review findings, and drive remediation from your terminal
or CI.

> **Status: pre-release.** The command surface and runtime are still being finalized
> (single-binary, Go/Rust candidate). Interfaces may change before the first tagged
> release. Watch [Releases](https://github.com/verifi-security-platform/verifi-cli/releases)
> for the first stable cut.

## Install

_Coming with the first release._ Distribution will be a single self-contained binary
(no runtime dependencies) plus package-manager taps.

## Usage

```
verifi scan <path>        # scan a project's dependencies
verifi status             # show findings for the current project
verifi login              # authenticate against a Verifi backend
```

Full command reference lands with the first release.

## How it fits together

The CLI talks only to the Verifi backend API — it never reaches into internal services
directly. You can run it against Verifi's hosted backend or a self-hosted enterprise
install.

```
  you ──▶ verifi CLI ──▶ Verifi backend API ──▶ detection · triage · remediation
```

## Contributing

Issues and pull requests are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md)
first, and note that contributions are accepted under the terms of the project license.

## Security

Found a vulnerability? **Do not open a public issue.** Follow the coordinated-disclosure
process in [SECURITY.md](SECURITY.md).

## License

Licensed under the [Apache License, Version 2.0](LICENSE).
Copyright © 2026 Verifi Security.
