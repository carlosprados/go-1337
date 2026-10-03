# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`go-1337` (module `github.com/carlosprados/go-1337`): converts text to and from leet speak.
One Go engine with three front ends: a Cobra CLI with a Bubble Tea live editor (binary `leet`),
and a browser demo that runs the engine compiled to WebAssembly.

## Commands

```bash
just build                    # ./leet with git version stamped via -ldflags
just check                    # gofmt check, go vet, go test -race (what CI runs)
go test -run TestDecode -v ./internal/leet   # single test
just web                      # wasm + npm install + hugo server (web demo, live reload)
just web-build                # wasm + npm ci + tsc type-check + hugo build into web/public
just demo                     # re-record docs/demo.gif; needs vhs >= 0.12.1 (0.12.0 silently writes no GIF)
just release-check            # goreleaser check + snapshot build into dist/
```

Version is injected into `github.com/carlosprados/go-1337/internal/cli.version`.

## Architecture

- `internal/leet` — the engine, no CLI or UI dependencies. Every letter has an ordered variant list (`defaultVariants`); index 0 is the **primary**, used by deterministic encoding. `Level` (basic/advanced/elite) gates which letters encode, via `minLevel`. `Alphabet` holds variants plus a decoding index of tokens sorted longest-first, then primary-first. `Decode` and `Detect` share `scan`. `config.go` loads YAML overrides (scalar or list per letter).
- `internal/cli` — Cobra commands built by constructor functions (`newEncodeCmd`, …) from `newRootCmd`, so tests build a fresh tree per run. `alphabet()` resolves `--map`, then the optional `$XDG_CONFIG_HOME/leet/map.yaml`. Root with no args opens the TUI only when stdin and stdout are TTYs; otherwise it prints help.
- `internal/tui` — Bubble Tea model. Random mode uses a fixed `seed` per render so output is stable while typing; `ctrl+s` bumps the seed. Panel sizes are inner sizes; the border adds 2 to each dimension.
- `cmd/leet` — the binary's `main`. `cmd/wasm` — `js && wasm` build exposing a global `leet` object (`encode(text, level, seed)`, `decode`, `detect`) and firing a `leet-ready` event.
- `web/` — Hugo site: `layouts/home.html`, `assets/ts/main.ts` (VanJS, bundled by Hugo `js.Build`), `assets/css/main.css` (Tailwind v4 via `css.TailwindCSS`, `@source` covers `layouts` and `assets/ts`). `web/static/leet.wasm` and `wasm_exec.js` are build outputs (gitignored), copied from `$(go env GOROOT)/lib/wasm/`.
- Releases: `.goreleaser.yaml` (binaries, completions, Homebrew cask that is skipped without `HOMEBREW_TAP_GITHUB_TOKEN`). Workflows: `go.yml` (checks), `release.yml` (on `v*` tags), `pages.yml` (web demo to GitHub Pages).

## Invariants — tests enforce them

- No variant may contain letters or whitespace (`validate`). This is what keeps plain words intact when decoding.
- Primaries are unique across letters, so `Decode(Encode(x, any level)) == lower(x)` for text without variant characters. `1` is the primary of `l` and a secondary of `i`: the only allowed duplicate.
- The CLI writes only converted text to stdout. Prompts go to stderr, and only when the input is a real terminal (`term.IsTerminal`; a `Stat` check would treat `/dev/null` as a TTY).
- Each command's `Long`/`Example` is the user documentation; update them, and the README, with behaviour changes.

## Gotchas

- `!` is the primary of `i`, so `!` in plain prose decodes as `i`. Digits and symbols in normal text are always read as leet; this is inherent, not a bug to fix with heuristics.
- `--random` output is not guaranteed to round-trip: alternatives can combine into another letter's sequence (`|_|_|` → `uj`).
- Web: JetBrains Mono ligatures are disabled on purpose, because they render `<|<` as an arrow. VanJS: the textarea is created once outside any derive and bound to the state object, and dynamic attributes use per-property derives (see the global VanJS rules).
