version := `git describe --tags --always --dirty 2>/dev/null || echo dev`

# List recipes
default:
    @just --list

# Build the leet binary with the version stamped in
build:
    go build -trimpath -ldflags "-s -w -X github.com/carlosprados/go-1337/internal/cli.version={{version}}" -o leet ./cmd/leet

# Format check, vet and race-enabled tests
check:
    test -z "$(gofmt -l .)" || { gofmt -l .; exit 1; }
    go vet ./...
    go test -race ./...

# Install leet into $GOBIN
install:
    go install -ldflags "-X github.com/carlosprados/go-1337/internal/cli.version={{version}}" ./cmd/leet

# Compile the converter to WebAssembly for the web demo
wasm:
    GOOS=js GOARCH=wasm go build -trimpath -ldflags "-s -w" -o web/static/leet.wasm ./cmd/wasm
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" web/static/

# Run the web demo locally with live reload
web: wasm
    cd web && npm install --no-fund --no-audit && hugo server

# Build the web demo into web/public
web-build: wasm
    cd web && npm ci && npx tsc -p . && hugo build --minify --environment production

# Record docs/demo.gif from docs/demo.tape (needs vhs, ttyd, ffmpeg)
demo: build
    PATH="$PWD:$PATH" vhs docs/demo.tape

# Validate the release config and build a local snapshot into dist/
release-check:
    goreleaser check
    goreleaser release --snapshot --clean
