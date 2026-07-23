# Verifi CLI

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)

Fix vulnerable dependencies without breaking your app.

Every scanner hands you a list of vulnerable dependencies. None of them tell you whether
fixing one will break your app. So the list sits there, or someone upgrades and loses an
afternoon to a broken build.

`verifi` looks at how your project actually uses each package and tells you which fixes are
safe, which ones need a second look, and why. You get a short, trustworthy list of changes
instead of a backlog you are afraid to touch. It is not another scanner. Its job starts
where detection stops.

Estate-wide response, blocking installs at the registry and coordinating fixes across
repositories, is the job of the wider [Verifi](https://verifisecurity.com) platform the CLI
plugs into.

> **Status: pre-release, built in the open.** The command surface is still being shaped and
> interfaces may change before the first tagged release. Watch
> [Releases](https://github.com/verifisecurity/verifi/releases) for the first stable cut.

## Install

_Coming with the first release._ Distribution will be a single self-contained binary with
no runtime dependencies.

## What you get

- **Focus on what matters.** Skip the vulnerabilities that never reach your code, and spend
  your time on the ones that do.
- **No surprise broken builds.** See what a fix will change in your project before you make
  it, with the reason it is safe or the catch to watch for.
- **Ship the fix with confidence.** Move faster on security, because you know what a change
  will do before you commit to it.

## Usage

```
verifi status <path>   # what needs fixing, and which fixes are safe to ship
verifi fix <path>      # apply a safe fix in your project
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
