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
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

var (
	wasmLoc  = flag.String("wasm-path", "", "path to the web app wasm file")
	port     = flag.Int("port", 7000, "default port to use")
	clientID = flag.String("google-client-id", "841378387526-qn3965fuan6b08os2pf33ul3lqr800dn.apps.googleusercontent.com", "Google Client ID")
)

// IsochronePanel component
type IsochronePanel struct {
	app.Compo
	address  string
	apiKey   string
	errorMsg string
}

func (h *IsochronePanel) Render() app.UI {
	return app.Div().Style("padding", "20px").Body(
		app.H2().Text("Isochrone Map (Panel 1)"),
		app.P().Text("Enter an address and ORS API key to see reachable area."),
		app.Div().Style("margin-bottom", "20px").Body(
			app.Input().
				Type("text").
				Value(h.address).
				Placeholder("Enter an address...").
				Style("width", "100%").
				Style("padding", "8px").
				Style("margin-bottom", "10px").
				OnChange(h.OnAddressChange),
			app.Input().
				Type("password").
				Value(h.apiKey).
				Placeholder("ORS API Key").
				Style("width", "100%").
				Style("padding", "8px").
				Style("margin-bottom", "10px").
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
			Style("height", "400px").
			Style("background-color", "#e0e0e0").
			Style("border", "1px solid #ccc").
			Style("border-radius", "4px"),
	)
}

func (h *IsochronePanel) OnMount(ctx app.Context) {
	ctx.Async(func() {
		app.Window().Call("loadMap", 51.505, -0.09)
	})
}

func (h *IsochronePanel) OnAddressChange(ctx app.Context, e app.Event) {
	h.address = ctx.JSSrc().Get("value").String()
}

func (h *IsochronePanel) OnAPIKeyChange(ctx app.Context, e app.Event) {
	h.apiKey = ctx.JSSrc().Get("value").String()
}

func (h *IsochronePanel) OnSearch(ctx app.Context, e app.Event) {
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
			isoReq.Header.Set("Accept", "application/json")

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

// GreetingPanel component
type GreetingPanel struct {
	app.Compo
	Token    string
	UserName string
	ErrorMsg string
}

func (g *GreetingPanel) Render() app.UI {
	u, _ := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	q := u.Query()
	q.Set("client_id", *clientID)
	q.Set("response_type", "token")
	q.Set("scope", "https://www.googleapis.com/auth/userinfo.profile")
	q.Set("state", "login")
	q.Set("redirect_uri", fmt.Sprintf("http://localhost:%d/foobar", *port))
	u.RawQuery = q.Encode()

	return app.Div().Style("padding", "20px").Body(
		app.H2().Text("Google Login (Panel 2)"),
		app.If(g.Token == "",
			app.Div().Body(
				app.P().Text("Login with Google to see a personalized greeting."),
				app.A().
					Text("Log In").
					Href(u.String()).
					Style("display", "inline-block").
					Style("padding", "10px 20px").
					Style("background-color", "#4285F4").
					Style("color", "white").
					Style("text-decoration", "none").
					Style("border-radius", "4px"),
			),
		).Else(
			app.Div().Body(
				app.If(g.UserName != "",
					app.H1().Text("Hello world, "+g.UserName),
				).ElseIf(g.ErrorMsg != "",
					app.P().Style("color", "red").Text(g.ErrorMsg),
				).Else(
					app.P().Text("Loading profile..."),
				),
				app.Button().
					Text("Logout").
					Style("padding", "8px 16px").
					Style("margin-top", "20px").
					OnClick(g.OnLogout),
			),
		),
	)
}

func (g *GreetingPanel) OnMount(ctx app.Context) {
	ctx.ObserveState("/login").OnChange(func() {
		if g.Token != "" && g.UserName == "" {
			g.fetchProfile(ctx)
		}
	}).Value(&g.Token)
	if g.Token != "" && g.UserName == "" {
		g.fetchProfile(ctx)
	}
}

func (g *GreetingPanel) fetchProfile(ctx app.Context) {
	ctx.Async(func() {
		req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) { g.ErrorMsg = "Failed to create request: " + err.Error() })
			return
		}
		req.Header.Set("Authorization", "Bearer "+g.Token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			ctx.Dispatch(func(ctx app.Context) { g.ErrorMsg = "Network error: " + err.Error() })
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			ctx.Dispatch(func(ctx app.Context) { g.ErrorMsg = fmt.Sprintf("API error (%d)", resp.StatusCode) })
			return
		}

		var profile struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&profile); err == nil {
			ctx.Dispatch(func(ctx app.Context) {
				g.UserName = profile.Name
				g.ErrorMsg = ""
			})
		} else {
			ctx.Dispatch(func(ctx app.Context) { g.ErrorMsg = "Failed to parse profile" })
		}
	})
}

func (g *GreetingPanel) OnLogout(ctx app.Context, e app.Event) {
	ctx.SetState("/login", "", app.Persist)
	g.Token = ""
	g.UserName = ""
}

// LoginCallback component
type LoginCallback struct {
	app.Compo
}

func (l *LoginCallback) Render() app.UI {
	return app.Main().Body(app.P().Text("Logging in..."))
}

func (l *LoginCallback) OnMount(ctx app.Context) {
	w := app.Window()
	v, _ := url.ParseQuery(w.URL().Fragment)
	re, _ := strconv.Atoi(v.Get("expires_in"))
	exp := time.Duration(re) * time.Second
	t := v.Get("access_token")

	if t != "" {
		ctx.SetState("/login", t, app.ExpiresIn(exp))
	}
	ctx.Navigate("/")
}

// GraphsPanel component
type GraphsPanel struct {
	app.Compo
}

func (p *GraphsPanel) Render() app.UI {
	return app.Div().Style("padding", "20px").Body(
		app.H2().Text("Example Graphs (Panel 3)"),
		app.Div().Style("display", "flex").Style("flex-wrap", "wrap").Style("gap", "20px").Body(
			app.Div().Style("width", "400px").Body(
				app.Canvas().ID("barChart"),
			),
			app.Div().Style("width", "400px").Body(
				app.Canvas().ID("lineChart"),
			),
			app.Div().Style("width", "300px").Body(
				app.Canvas().ID("doughnutChart"),
			),
			app.Div().Style("width", "600px").Style("height", "500px").ID("plot3d"),
		),
	)
}

func (p *GraphsPanel) OnMount(ctx app.Context) {
	ctx.Async(func() {
		for {
			if !app.Window().Get("Chart").IsUndefined() && !app.Window().Get("Plotly").IsUndefined() {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		app.Window().Call("drawExampleCharts")
		app.Window().Call("draw3DChart")
	})
}

// Root component
type Root struct {
	app.Compo
	currentApp string
}

func (r *Root) Render() app.UI {
	isIsochrone := r.currentApp == "isochrone" || r.currentApp == ""
	isGreeting := r.currentApp == "greeting"
	isGraphs := r.currentApp == "graphs"

	return app.Main().Style("font-family", "sans-serif").Body(
		app.Nav().Style("display", "flex").Style("background-color", "#333").Style("padding", "10px").Body(
			app.Button().Text("Isochrone Map").
				Style("margin-right", "10px").
				Style("background-color", map[bool]string{true: "#555", false: "#333"}[isIsochrone]).
				Style("color", "white").
				Style("border", "none").
				Style("padding", "10px 20px").
				Style("cursor", "pointer").
				Style("border-radius", "4px").
				OnClick(r.showIsochrone),
			app.Button().Text("Google Login").
				Style("margin-right", "10px").
				Style("background-color", map[bool]string{true: "#555", false: "#333"}[isGreeting]).
				Style("color", "white").
				Style("border", "none").
				Style("padding", "10px 20px").
				Style("cursor", "pointer").
				Style("border-radius", "4px").
				OnClick(r.showGreeting),
			app.Button().Text("Example Graphs").
				Style("background-color", map[bool]string{true: "#555", false: "#333"}[isGraphs]).
				Style("color", "white").
				Style("border", "none").
				Style("padding", "10px 20px").
				Style("cursor", "pointer").
				Style("border-radius", "4px").
				OnClick(r.showGraphs),
		),
		app.Div().Style("display", map[bool]string{true: "block", false: "none"}[isIsochrone]).Body(
			&IsochronePanel{},
		),
		app.Div().Style("display", map[bool]string{true: "block", false: "none"}[isGreeting]).Body(
			&GreetingPanel{},
		),
		app.Div().Style("display", map[bool]string{true: "block", false: "none"}[isGraphs]).Body(
			&GraphsPanel{},
		),
	)
}

func (r *Root) showIsochrone(ctx app.Context, e app.Event) {
	r.currentApp = "isochrone"
}

func (r *Root) showGreeting(ctx app.Context, e app.Event) {
	r.currentApp = "greeting"
}

func (r *Root) showGraphs(ctx app.Context, e app.Event) {
	r.currentApp = "graphs"
}

func main() {
	app.Route("/", &Root{})
	app.Route("/foobar", &LoginCallback{})

	app.RunWhenOnBrowser()

	flag.Parse()

	h := &app.Handler{
		Title: "Isochrone & Google Login WebApp",
		Styles: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.css",
		},
		Scripts: []string{
			"https://unpkg.com/leaflet@1.9.4/dist/leaflet.js",
			"https://cdn.jsdelivr.net/npm/chart.js",
			"https://cdn.plot.ly/plotly-2.32.0.min.js",
		},
		RawHeaders: []string{
			`
			<script>
			function draw3DChart() {
				var z_data = [];
				for(var i=0;i<25;i++) {
					var row = [];
					for(var j=0;j<25;j++) {
						row.push(Math.sin(i/3.0) * Math.cos(j/3.0));
					}
					z_data.push(row);
				}
				var data = [{
					z: z_data,
					type: 'contour'
				}];
				var layout = {
					title: '2D Contour Plot (WebGL Fallback)',
					autosize: true,
					margin: {l: 50, r: 50, b: 50, t: 50}
				};
				Plotly.newPlot('plot3d', data, layout);
			}

			function drawExampleCharts() {
				if (window.myBarChart) window.myBarChart.destroy();
				if (window.myLineChart) window.myLineChart.destroy();
				if (window.myDoughnutChart) window.myDoughnutChart.destroy();

				const barCtx = document.getElementById('barChart');
				if (barCtx) {
					window.myBarChart = new Chart(barCtx, {
						type: 'bar',
						data: {
							labels: ['Red', 'Blue', 'Yellow', 'Green', 'Purple', 'Orange'],
							datasets: [{
								label: '# of Votes',
								data: [12, 19, 3, 5, 2, 3],
								backgroundColor: [
									'rgba(255, 99, 132, 0.2)', 'rgba(54, 162, 235, 0.2)', 'rgba(255, 206, 86, 0.2)',
									'rgba(75, 192, 192, 0.2)', 'rgba(153, 102, 255, 0.2)', 'rgba(255, 159, 64, 0.2)'
								],
								borderColor: [
									'rgba(255, 99, 132, 1)', 'rgba(54, 162, 235, 1)', 'rgba(255, 206, 86, 1)',
									'rgba(75, 192, 192, 1)', 'rgba(153, 102, 255, 1)', 'rgba(255, 159, 64, 1)'
								],
								borderWidth: 1
							}]
						},
						options: { scales: { y: { beginAtZero: true } } }
					});
				}

				const lineCtx = document.getElementById('lineChart');
				if (lineCtx) {
					window.myLineChart = new Chart(lineCtx, {
						type: 'line',
						data: {
							labels: ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul'],
							datasets: [{
								label: 'Monthly Sales',
								data: [65, 59, 80, 81, 56, 55, 40],
								fill: false,
								borderColor: 'rgb(75, 192, 192)',
								tension: 0.1
							}]
						}
					});
				}

				const doughnutCtx = document.getElementById('doughnutChart');
				if (doughnutCtx) {
					window.myDoughnutChart = new Chart(doughnutCtx, {
						type: 'doughnut',
						data: {
							labels: ['Download Sales', 'In-Store Sales', 'Mail-Order Sales'],
							datasets: [{
								label: 'Sales Distribution',
								data: [300, 50, 100],
								backgroundColor: ['rgb(255, 99, 132)', 'rgb(54, 162, 235)', 'rgb(255, 205, 86)'],
								hoverOffset: 4
							}]
						}
					});
				}
			}

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
