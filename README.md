# Verifi CLI

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

The open-source command-line tool for finding and fixing risky dependencies, from
[Verifi](https://verifisecurity.com).

`verifi` is a small, cross-platform CLI. Point it at a project and it inspects your
dependencies for known vulnerabilities, end-of-life packages, and malicious releases,
then helps you fix what has a fix. It runs where you already work: your terminal and CI.

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
