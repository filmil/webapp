package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

var (
	wasmLoc = flag.String("wasm-path", "", "path to the web app wasm file")
	port    = flag.Int("port", 7000, "default port to use")
)

type Hello struct {
	app.Compo
}

func (h *Hello) Render() app.UI {
	return app.Main().Body(
		app.H1().Text("Hello, World!"),
		app.P().Text("This is a simple go-app.dev app built with Bazel."),
	)
}

func main() {
	app.Route("/", &Hello{})

	app.RunWhenOnBrowser()

	flag.Parse()

	h := &app.Handler{
		Title: "Hello World Go-App",
	}

	mux := http.NewServeMux()

	if *wasmLoc != "" {
		mux.HandleFunc("/web/app.wasm", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, *wasmLoc)
		})
	}

	mux.Handle("/", h)

	log.Printf("Starting local server on http://localhost:%v\n", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", *port), mux))
}
