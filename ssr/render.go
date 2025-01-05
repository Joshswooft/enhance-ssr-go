package ssr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	extism "github.com/extism/go-sdk"
)

type ElementName string
type ElementFileContents string

type ElementContentsMap map[ElementName]ElementFileContents

type Payload struct {
	// Markup is the html contents of the page
	Markup string `json:"markup"`
	// Elements is a map containing the name of the element as its key and its file contents
	Elements     ElementContentsMap     `json:"elements"`
	InitialState map[string]interface{} `json:"initialState"`
}

// Marshal formats the JSON output with indentation and without escaping HTML characters.
// It mimics json.MarshalIndent but without HTML escaping.
func (p *Payload) MarshalJSON() ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(p)
	if err != nil {
		return nil, err
	}
	// Trim the trailing newline added by Encode
	return bytes.TrimRight(buffer.Bytes(), "\n"), nil
}

type RenderResult struct {
	// Document is the html of the rendered content
	Document string `json:"document"`
}

// Render custom elements from given markup and state using the enhance wasm
func Render(ctx context.Context, payload Payload) (*RenderResult, error) {
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmFile{
				Path: "./enhance-ssr.wasm",
			},
		},
	}

	config := extism.PluginConfig{
		EnableWasi: true,
	}
	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %v", err)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create payload, err=%v, payload=%v", err, payload)
	}

	exit, out, err := plugin.Call("ssr", payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("plugin call failed: %v, exit code: %d", err, exit)
	}

	var result RenderResult

	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to parse render results: %v", err)
	}

	return &result, nil
}
