//go:build js && wasm

// Command wasm exposes the leet converter to JavaScript for the web demo.
//
//	GOOS=js GOARCH=wasm go build -o web/static/leet.wasm ./cmd/wasm
//
// It registers a global "leet" object with encode(text, level, seed),
// decode(text) and detect(text). A seed of 0 means primary variants only.
package main

import (
	"math/rand/v2"
	"syscall/js"

	"github.com/carlosprados/go-1337/internal/leet"
)

func main() {
	a := leet.Default()
	js.Global().Set("leet", js.ValueOf(map[string]any{
		"encode": js.FuncOf(func(_ js.Value, args []js.Value) any {
			level, err := leet.ParseLevel(args[1].String())
			if err != nil {
				level = leet.Elite
			}
			opts := leet.EncodeOptions{Level: level}
			if seed := uint64(args[2].Int()); seed != 0 {
				opts.Rand = rand.New(rand.NewPCG(seed, seed))
			}
			return a.Encode(args[0].String(), opts)
		}),
		"decode": js.FuncOf(func(_ js.Value, args []js.Value) any {
			return a.Decode(args[0].String())
		}),
		"detect": js.FuncOf(func(_ js.Value, args []js.Value) any {
			s := a.Detect(args[0].String())
			return map[string]any{"ratio": s.Ratio, "verdict": s.Verdict(), "decoded": s.Decoded}
		}),
		"levels": js.FuncOf(func(js.Value, []js.Value) any {
			out := make([]any, 0, 3)
			for _, l := range leet.Levels() {
				out = append(out, l)
			}
			return out
		}),
	}))
	js.Global().Call("dispatchEvent", js.Global().Get("Event").New("leet-ready"))
	select {}
}
