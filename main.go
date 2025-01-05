package main

import (
	"embed"
	"enhance/enhance-ssr-go/ssr"
	"fmt"
	"net/http"
	"path/filepath"
)

func main() {
	http.HandleFunc("/", handleRequest)
	fmt.Println("Server starting on http://localhost:8080 ...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	markup := `<my-header>My custom header</my-header>`
	elements := readElementsFromEmbed(customElements)
	initialState := make(map[string]interface{})
	data := ssr.Payload{
		Markup:       markup,
		Elements:     elements,
		InitialState: initialState,
	}
	rendered, err := ssr.Render(r.Context(), data)
	if err != nil {
		http.Error(w, "Failed to render document", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, rendered.Document)
}

//go:embed elements
var customElements embed.FS

func readElementsFromEmbed(fs embed.FS) ssr.ElementContentsMap {
	elements := make(ssr.ElementContentsMap)
	entries, err := fs.ReadDir("elements")
	if err != nil {
		fmt.Printf("Error reading embedded directory: %s\n", err)
		return elements
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			content, err := fs.ReadFile("elements/" + entry.Name())
			if err != nil {
				fmt.Printf("Error reading embedded file %s: %s\n", entry.Name(), err)
				continue
			}
			key := filepath.Base(entry.Name())
			ext := filepath.Ext(key)
			keyWithoutExt := key[:len(key)-len(ext)]
			elements[ssr.ElementName(keyWithoutExt)] = ssr.ElementFileContents(content)
		}
	}
	return elements
}
