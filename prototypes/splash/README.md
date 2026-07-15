# `verifi` welcome splash — design prototype

A runnable mock of the CLI's first-run welcome: muted boxes of varying sizes
fall down a blackish gradient (the website hero motif), then fade out as the
solid **VERIFI** wordmark and a `coming soon · beta` line settle in on top.

This is a **design prototype**, not the real CLI — stdlib Go only, no deps, no
install pipeline. It exists so we can see and tune the look before wiring up a
real binary + release.

## Run it

```sh
cd prototypes/splash
go run .            # play the animation once
go run . --loop     # replay forever while tuning (Ctrl-C to quit)
go run . --static   # just the settled banner, no animation
```

Needs a truecolor terminal (VS Code's terminal, iTerm2, and modern
Terminal.app all qualify). When output isn't a TTY (piped/CI) it prints the
static banner automatically.

## Tuning

All the knobs are the `TUNABLES` block at the top of `main.go`:

- `fallSeconds` / `settleSeconds` / `holdSeconds` — timing of the three phases.
- `spawnChance`, `blockDim`, `bottomFade` — density and fade of the falling boxes.
- `palette`, `bgTop`/`bgBottom`, `inkWord`/`inkTeal` — colours (palette is the
  web hero's, dimmed by `blockDim`).
- box sizes live in `spawnBox()` (bias toward small, occasional larger ones).

## Next step

When the look is right, this ports into the real `verifi` binary as the
`welcome` path (Go + Bubble Tea for the production version), fronted by a real
`install.sh` (OS/arch detect + checksum) publishing to GitHub Releases.
