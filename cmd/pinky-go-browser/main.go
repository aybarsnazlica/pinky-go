//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	pinky "github.com/aybarsnazlica/pinky-go"
)

func main() {
	runtime := js.Global().Get("Object").New()
	runtime.Set("sampleProgram", pinky.SampleProgram)
	runtime.Set("runSource", js.FuncOf(func(this js.Value, args []js.Value) any {
		source := ""
		if len(args) > 0 {
			source = args[0].String()
		}

		includeDebug := false
		if len(args) > 1 {
			includeDebug = args[1].Bool()
		}

		result := pinky.RunSource(source, includeDebug)
		encoded, err := json.Marshal(result)
		if err != nil {
			fallback, fallbackErr := json.Marshal(map[string]any{
				"success":      false,
				"source":       source,
				"tokens":       []string{},
				"ast":          "",
				"output":       "",
				"errorType":    "internal",
				"errorMessage": err.Error(),
				"errorLine":    0,
			})
			if fallbackErr != nil {
				return `{"success":false,"source":"","tokens":[],"ast":"","output":"","errorType":"internal","errorMessage":"Failed to encode result.","errorLine":0}`
			}
			return string(fallback)
		}

		return string(encoded)
	}))

	js.Global().Set("__pinkyGoRuntime", runtime)

	select {}
}
