package main

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type Hello struct {
	app.Compo
}

func (h *Hello) Render() app.UI {
	return app.Main().Body(
		app.H1().Text("Hello, World!"),
		app.P().Text("This is a simple go-app.dev app built with Bazel."),
	)
}

func init() {
	app.Route("/", &Hello{})
}
