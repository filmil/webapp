package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

var (
	wasmLoc = flag.String("wasm-path", "", "path to the web app wasm file")
	port    = flag.Int("port", 7000, "default port to use")
)

type Hello struct {
	app.Compo
	address  string
	errorMsg string
}

func (h *Hello) Render() app.UI {
	return app.Main().Style("padding", "20px").Style("font-family", "sans-serif").Body(
		app.H1().Text("OpenStreetMap Geocoder"),
		app.P().Text("Enter an address to zoom to the location."),
		app.Div().Style("margin-bottom", "20px").Body(
			app.Input().
				Type("text").
				Value(h.address).
				Placeholder("Enter an address...").
				Style("width", "300px").
				Style("padding", "8px").
				Style("margin-right", "10px").
				OnChange(h.OnInputChange),
			app.Button().
				Text("Search Address").
				Style("padding", "8px 16px").
				OnClick(h.OnSearch),
		),
		app.If(h.errorMsg != "",
			app.Div().Text(h.errorMsg).Style("color", "red").Style("margin-bottom", "10px"),
		),
		app.Div().
			ID("map").
			Style("width", "100%").
			Style("height", "500px").
			Style("background-color", "#e0e0e0").
			Style("border", "1px solid #ccc").
			Style("border-radius", "4px"),
	)
}

func (h *Hello) OnMount(ctx app.Context) {
	// Initialize map to a default location (e.g., London) when the component mounts
	ctx.Async(func() {
		app.Window().Call("loadMap", 51.505, -0.09)
	})
}

func (h *Hello) OnInputChange(ctx app.Context, e app.Event) {
	h.address = ctx.JSSrc().Get("value").String()
}

func (h *Hello) OnSearch(ctx app.Context, e app.Event) {
	if h.address == "" {
		return
	}
	h.errorMsg = "Searching..."

	ctx.Async(func() {
		query := url.QueryEscape(h.address)
		urlStr := "https://nominatim.openstreetmap.org/search?format=json&q=" + query

		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) { h.errorMsg = err.Error() })
			return
		}
		// Nominatim requires a valid User-Agent
		req.Header.Set("User-Agent", "GeminiCLI-DemoApp/1.0")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) { h.errorMsg = "Network error: " + err.Error() })
			return
		}
		defer resp.Body.Close()

		var results []struct {
			Lat string `json:"lat"`
			Lon string `json:"lon"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
			ctx.Dispatch(func(ctx app.Context) { h.errorMsg = "Failed to parse response" })
			return
		}

		if len(results) == 0 {
			ctx.Dispatch(func(ctx app.Context) { h.errorMsg = "Address not found" })
			return
		}

		lat, _ := strconv.ParseFloat(results[0].Lat, 64)
		lon, _ := strconv.ParseFloat(results[0].Lon, 64)

		ctx.Dispatch(func(ctx app.Context) {
			h.errorMsg = ""
			app.Window().Call("loadMap", lat, lon)
		})
	})
}

func main() {
	app.Route("/", &Hello{})

	app.RunWhenOnBrowser()

	flag.Parse()

	h := &app.Handler{
		Title: "Map WebApp",
		Styles: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.css",
		},
		Scripts: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.js",
		},
		RawHeaders: []string{
			`
			<script>
			var myMap;
			function loadMap(lat, lon) {
				if (typeof L === 'undefined') {
					setTimeout(function() { loadMap(lat, lon); }, 100);
					return;
				}
				var mapEl = document.getElementById('map');
				if (!mapEl) {
					setTimeout(function() { loadMap(lat, lon); }, 100);
					return;
				}
				if (!myMap) {
					myMap = L.map('map').setView([lat, lon], 13);
					L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
						maxZoom: 19,
						attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
					}).addTo(myMap);
				} else {
					myMap.setView([lat, lon], 13);
				}
			}
			</script>
			`,
		},
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
