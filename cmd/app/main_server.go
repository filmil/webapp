package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

//go:embed app_bundled.wasm
var wasmFS embed.FS

var (
	port = flag.Int("port", 7000, "default port to use")
)

func main() {
	flag.Parse()

	h := &app.Handler{
		Title: "Hello World Go-App",
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/web/app.wasm", func(w http.ResponseWriter, r *http.Request) {
		wasmFile, err := wasmFS.ReadFile("app_bundled.wasm")
		if err != nil {
			http.Error(w, "wasm not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/wasm")
		w.Write(wasmFile)
	})

	mux.Handle("/", h)

	log.Printf("Starting local server on http://localhost:%v\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), mux))
}
