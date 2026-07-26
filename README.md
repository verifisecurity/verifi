# Verifi CLI

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Fix vulnerable dependencies without breaking your app.

Every scanner hands you a list of vulnerable dependencies. None of them tell you whether
fixing one will break your app. So the list sits there, or someone upgrades and loses an
afternoon to a broken build.

`verifi` looks at how your project actually uses each package and, for each vulnerable one,
tells you the version to move to, what it clears, and what it has not checked. You get a
short, honest list of fixes instead of a backlog you are afraid to touch. It is not another
scanner. Its job starts where detection stops.

> **Status: v0.1.0, pre-1.0.** The read-only core is here: it scans a project and reasons
> about each fix. The command surface may still change before 1.0. See
> [Releases](https://github.com/verifisecurity/verifi/releases).

## Install

<!-- BEGIN INSTALL -->
**macOS and Linux**

```sh
curl -fsSL https://raw.githubusercontent.com/verifisecurity/verifi/main/install.sh | sh
```

**Windows** (PowerShell)

```powershell
irm https://raw.githubusercontent.com/verifisecurity/verifi/main/install.ps1 | iex
```
<!-- END INSTALL -->

A single self-contained binary with no runtime dependencies. Prefer to do it yourself? Grab a
build from [Releases](https://github.com/verifisecurity/verifi/releases).

## Usage

Scan a project:

```sh
verifi scan path/to/project
```

The first scan needs the OSV advisory database. If it is not on the machine yet,
`scan` says so, and `--download` fetches it into `~/.verifi/osv` for every project
to share:

```sh
verifi scan path/to/project --download
```

You get, per vulnerable package, whether your code imports it, the version to upgrade to, what
that clears, and the limits of what has been checked.

<!-- BEGIN COMMANDS -->
```
verifi scan <path>  Report what is vulnerable and the fix for each one
verifi fix <path>   Apply the recommended fix, or preview it (--apply, --db, --offline)
verifi version      Print the version
verifi help         Show this help
```
<!-- END COMMANDS -->

## Supported ecosystems

<!-- BEGIN ECOSYSTEMS -->
```
npm  package-lock.json
```

More ecosystems are on the way.
<!-- END ECOSYSTEMS -->

## What you get

- **Focus on what matters.** See which vulnerable packages your code actually imports, so you
  spend time on the ones that reach your app.
- **A concrete fix, not just a warning.** For each one, the version to move to and what it
  clears, checked against the registry so it is a version that really exists.
- **Honest about the limits.** Every fix says what has been verified and what has not, so you
  are never told more confidence than we have.

## Contributing

Issues and pull requests are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md)
first, and note that contributions are accepted under the terms of the project license.

## Security

Found a vulnerability? **Do not open a public issue.** Follow the coordinated-disclosure
process in [SECURITY.md](SECURITY.md).

## License

Licensed under the [Apache License, Version 2.0](LICENSE).
Copyright © 2026 Verifi Security.
