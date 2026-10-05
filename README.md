# leet — 1337 speak converter

[![Go](https://github.com/carlosprados/go-1337/actions/workflows/go.yml/badge.svg)](https://github.com/carlosprados/go-1337/actions/workflows/go.yml)
[![Release](https://img.shields.io/github/v/release/carlosprados/go-1337)](https://github.com/carlosprados/go-1337/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/carlosprados/go-1337.svg)](https://pkg.go.dev/github.com/carlosprados/go-1337)
[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

Convert text to `1337` speak and back: from the command line, in pipes, in a live
terminal editor, or **[in the browser](https://carlosprados.github.io/go-1337/)**.

![demo](docs/demo.gif)

## Install

### Download a binary

Each link always points to the latest release:

| OS | x86-64 | ARM64 |
|---|---|---|
| Linux | [leet_linux_amd64.tar.gz](https://github.com/carlosprados/go-1337/releases/latest/download/leet_linux_amd64.tar.gz) | [leet_linux_arm64.tar.gz](https://github.com/carlosprados/go-1337/releases/latest/download/leet_linux_arm64.tar.gz) |
| macOS | [leet_darwin_amd64.tar.gz](https://github.com/carlosprados/go-1337/releases/latest/download/leet_darwin_amd64.tar.gz) | [leet_darwin_arm64.tar.gz](https://github.com/carlosprados/go-1337/releases/latest/download/leet_darwin_arm64.tar.gz) (Apple Silicon) |
| Windows | [leet_windows_amd64.zip](https://github.com/carlosprados/go-1337/releases/latest/download/leet_windows_amd64.zip) | [leet_windows_arm64.zip](https://github.com/carlosprados/go-1337/releases/latest/download/leet_windows_arm64.zip) |

Or in one line on Linux and macOS, installing to `~/.local/bin`:

```bash
os=$(uname -s | tr A-Z a-z); arch=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
curl -sL "https://github.com/carlosprados/go-1337/releases/latest/download/leet_${os}_${arch}.tar.gz" \
  | tar -xz -C ~/.local/bin leet
```

`checksums.txt` on the [releases page](https://github.com/carlosprados/go-1337/releases) has the SHA-256 of every archive.
On macOS, a browser download is quarantined: run `xattr -d com.apple.quarantine leet` once.

### With Go

```bash
go install github.com/carlosprados/go-1337/cmd/leet@latest
```

Archives include shell completions; `leet completion --help` generates them too.

## Usage

| Command | Aliases | What it does |
|---|---|---|
| `leet` | | Live editor (when run in a terminal with no arguments) |
| `leet encode [text...]` | `to`, `to1337` | Plain text → 1337 |
| `leet decode [text...]` | `from`, `from1337` | 1337 → plain text |
| `leet detect [text...]` | `score` | How much of a text is 1337, and its decoded form |
| `leet table` | `map`, `alphabet` | Every letter, its level, primary variant and alternatives |
| `leet share [text...]` | | Print a web link that decrypts your message on screen |
| `leet tui` | `live`, `play` | Live editor, explicitly |
| `leet completion <shell>` | | Completion script for bash, zsh, fish, powershell |

Every command explains itself: `leet --help`, `leet <command> --help`.

Text comes from the arguments or, with none, from standard input line by line. Only the
converted text goes to stdout, so `leet` composes in pipes:

```bash
$ leet encode Hack the planet
#4<|< 7#3 |*14|\|37

$ leet encode --level basic Hack the planet
H4ck 7h3 p14n37

$ leet encode --random --seed 42 Hello      # --seed makes random output reproducible
|-|&11[]

$ leet encode 'leet speak' | leet decode
leet speak

$ leet detect 'h4ck th3 p14n37'
score    46% (mostly 1337)
decoded  hack the planet

$ leet encode -c "goes to the clipboard too"

$ leet decode --animate '|*455\/\/0|2|) 4<<3|*73|)'   # hacker-movie reveal
password accepted
```

`--animate` (`-a`) on `encode` and `decode` scrambles each line and settles it left to
right. It only plays when stdout is a terminal, so pipes always get plain text.

Quote leet input in the shell: many sequences contain `\ | < > * $`.

### Levels

| Level | Converts |
|---|---|
| `basic` | `a e l o s t` — the classic `1337` letters |
| `advanced` | basic + `b c g h i k z` |
| `elite` (default) | the whole alphabet |

Each letter has a **primary** variant, used by default, and alternatives that `--random`
picks from. `leet table` shows them all.

### Live editor

Run `leet` in a terminal. It converts as you type: side by side on wide terminals, stacked on narrow ones.
Switching mode, level or randomness replays the decrypting animation on the output.

| Key | Action |
|---|---|
| `tab` | switch encode / decode |
| `ctrl+l` | cycle level |
| `ctrl+r` | toggle random variants |
| `ctrl+s` | reshuffle random variants |
| `ctrl+y` | copy output to the clipboard |
| `esc` | quit |

### Share a secret message

`leet share` prints a link to the web demo carrying your message **in 1337**. Whoever
opens it watches it decrypt in the browser:

```bash
$ leet share "meet me at the usual place"
https://carlosprados.github.io/go-1337/#d=%2F%5C%2F%5C337+%2F%5C%2F%5C3+47+7%233+%7C_%7C5%7C_%7C41+%7C%2A14%3C3
```

The **share** button in the web demo does the same. The link holds only the leet, never the
plain text, and is decoded with the built-in alphabet (`--map` does not travel with it).

### Custom alphabet

Override any letter with a YAML file, passed with `--map` or picked up automatically from
`~/.config/leet/map.yaml` (`$XDG_CONFIG_HOME/leet/map.yaml`):

```yaml
a: "@"            # a single variant
e: ["3", "&"]     # primary first, then alternatives for --random
```

Variants must not contain letters or spaces. That rule keeps plain words intact when
decoding: a variant like `ph` would turn "phone" into "fone".

## How decoding works, and its limits

Decoding scans the input with longest-match-first over every variant of every letter. It is
deterministic and reads any level or random output. On a tie, the primary variant wins.

- **Primaries round-trip**: `decode(encode(x))` gives back `x` in lowercase, as long as `x`
  contains no characters that are themselves variants.
- **Symbols and digits in plain text are read as leet**: `Hi!` decodes as `hii`, `4` as `a`.
  `1` is the primary of `l`, so it decodes as `l` even when it stood for an `i`.
- **Random output may be ambiguous**: alternatives can combine into another letter's
  sequence (`lu` → `|_|_|` → `uj`).
- Case is not recovered.

## Web demo

`web/` is a [Hugo](https://gohugo.io) + [VanJS](https://vanjs.org) + Tailwind page that runs
the same Go converter and animation compiled to WebAssembly (`cmd/wasm`). It opens `#d=` share
links in decode mode, and respects `prefers-reduced-motion`. GitHub Pages publishes it on
every push to `main`.

## Development

Recipes live in the [`justfile`](justfile):

```bash
just build          # ./leet with the git version stamped in
just check          # gofmt, go vet, go test -race
just web            # web demo with live reload (needs Hugo extended and npm)
just demo           # re-record docs/demo.gif (needs vhs >= 0.12.1, ttyd, ffmpeg)
just release-check  # validate .goreleaser.yaml and build a local snapshot into dist/
```

Releases: push a `vX.Y.Z` tag and GoReleaser publishes binaries and checksums.

## License

MIT. See [LICENSE](LICENSE).
