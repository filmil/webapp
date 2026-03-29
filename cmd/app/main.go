package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

var (
	wasmLoc = flag.String("wasm-path", "", "path to the web app wasm file")
	port    = flag.Int("port", 7000, "default port to use")
)

type Hello struct {
	app.Compo
	address  string
	apiKey   string
	errorMsg string
}

func (h *Hello) Render() app.UI {
	return app.Main().Style("padding", "20px").Style("font-family", "sans-serif").Body(
		app.H1().Text("Isochrone Map (Reachable Area)"),
		app.P().Text("Enter an address and your OpenRouteService API key to see the reachable area by car at 1, 5, 10, 15, 20, 30, 40, and 60 minutes."),
		app.Div().Style("margin-bottom", "20px").Body(
			app.Input().
				Type("text").
				Value(h.address).
				Placeholder("Enter an address...").
				Style("width", "250px").
				Style("padding", "8px").
				Style("margin-right", "10px").
				OnChange(h.OnAddressChange),
			app.Input().
				Type("password").
				Value(h.apiKey).
				Placeholder("ORS API Key").
				Style("width", "200px").
				Style("padding", "8px").
				Style("margin-right", "10px").
				OnChange(h.OnAPIKeyChange),
			app.Button().
				Text("Search & Compute Area").
				Style("padding", "8px 16px").
				OnClick(h.OnSearch),
		),
		app.If(h.errorMsg != "",
			app.Div().Text(h.errorMsg).Style("color", "red").Style("margin-bottom", "10px"),
		),
		app.Div().
			ID("map").
			Style("width", "100%").
			Style("height", "600px").
			Style("background-color", "#e0e0e0").
			Style("border", "1px solid #ccc").
			Style("border-radius", "4px"),
	)
}

func (h *Hello) OnMount(ctx app.Context) {
	ctx.Async(func() {
		app.Window().Call("loadMap", 51.505, -0.09)
	})
}

func (h *Hello) OnAddressChange(ctx app.Context, e app.Event) {
	h.address = ctx.JSSrc().Get("value").String()
}

func (h *Hello) OnAPIKeyChange(ctx app.Context, e app.Event) {
	h.apiKey = ctx.JSSrc().Get("value").String()
}

func (h *Hello) OnSearch(ctx app.Context, e app.Event) {
	if h.address == "" {
		return
	}
	h.errorMsg = "Searching location..."

	ctx.Async(func() {
		// 1. Geocode
		query := url.QueryEscape(h.address)
		urlStr := "https://nominatim.openstreetmap.org/search?format=json&q=" + query

		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) { h.errorMsg = err.Error() })
			return
		}
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
			ctx.Dispatch(func(ctx app.Context) { h.errorMsg = "Failed to parse geocoding response" })
			return
		}

		if len(results) == 0 {
			ctx.Dispatch(func(ctx app.Context) { h.errorMsg = "Address not found" })
			return
		}

		lat, _ := strconv.ParseFloat(results[0].Lat, 64)
		lon, _ := strconv.ParseFloat(results[0].Lon, 64)

		ctx.Dispatch(func(ctx app.Context) {
			app.Window().Call("loadMap", lat, lon)
		})

		// 2. Fetch Isochrones if api key is provided
		if h.apiKey != "" {
			ctx.Dispatch(func(ctx app.Context) { h.errorMsg = "Computing reachable areas..." })

			reqBody := fmt.Sprintf(`{"locations":[[%f,%f]],"range":[60,300,600,900,1200,1800,2400,3600]}`, lon, lat)

			isoReq, err := http.NewRequest("POST", "https://api.openrouteservice.org/v2/isochrones/driving-car", strings.NewReader(reqBody))
			if err != nil {
				ctx.Dispatch(func(ctx app.Context) { h.errorMsg = "Isochrone request error: " + err.Error() })
				return
			}
			isoReq.Header.Set("Authorization", h.apiKey)
			isoReq.Header.Set("Content-Type", "application/json; charset=utf-8")
			isoReq.Header.Set("Accept", "application/json, application/geo+json, application/gpx+xml, img/png; charset=utf-8")

			isoResp, err := http.DefaultClient.Do(isoReq)
			if err != nil {
				ctx.Dispatch(func(ctx app.Context) { h.errorMsg = "Isochrone network error: " + err.Error() })
				return
			}
			defer isoResp.Body.Close()

			bodyBytes, _ := io.ReadAll(isoResp.Body)
			if isoResp.StatusCode != 200 {
				ctx.Dispatch(func(ctx app.Context) {
					h.errorMsg = fmt.Sprintf("Isochrone API error (%d): %s", isoResp.StatusCode, string(bodyBytes))
				})
				return
			}

			geoJsonStr := string(bodyBytes)
			ctx.Dispatch(func(ctx app.Context) {
				h.errorMsg = ""
				app.Window().Call("drawIsochrone", geoJsonStr)
			})
		} else {
			ctx.Dispatch(func(ctx app.Context) {
				h.errorMsg = "Please provide an OpenRouteService API Key to compute the reachable area."
			})
		}
	})
}

func main() {
	app.Route("/", &Hello{})

	app.RunWhenOnBrowser()

	flag.Parse()

	h := &app.Handler{
		Title: "Isochrone Map WebApp",
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
			var marker;
			var polygonLayer;

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
						attribution: '&copy; OpenStreetMap'
					}).addTo(myMap);
				} else {
					myMap.setView([lat, lon], 13);
				}

				if (marker) myMap.removeLayer(marker);
				marker = L.marker([lat, lon]).addTo(myMap);
			}

			function drawIsochrone(geoJsonStr) {
				if (!myMap) return;
				if (polygonLayer) myMap.removeLayer(polygonLayer);

				try {
					var geojson = JSON.parse(geoJsonStr);
					polygonLayer = L.geoJSON(geojson, {
						style: function (feature) {
							// Color code based on range value (in seconds)
							var value = feature.properties.value;
							var color = '#800026';
							if (value <= 60) color = '#ffffcc';
							else if (value <= 300) color = '#ffeda0';
							else if (value <= 600) color = '#fed976';
							else if (value <= 900) color = '#feb24c';
							else if (value <= 1200) color = '#fd8d3c';
							else if (value <= 1800) color = '#fc4e2a';
							else if (value <= 2400) color = '#e31a1c';
							else color = '#b10026';

							return {
								color: color,
								weight: 1,
								fillOpacity: 0.4
							};
						}
					}).addTo(myMap);
					myMap.fitBounds(polygonLayer.getBounds());
				} catch (e) {
					console.error("Error parsing/drawing GeoJSON", e);
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
