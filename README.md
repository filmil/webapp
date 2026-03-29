# Webapp

This project was created with the help of Gemini CLI.

## Status

[![CI](https://github.com/filmil/webapp/actions/workflows/ci.yml/badge.svg)](https://github.com/filmil/webapp/actions/workflows/ci.yml)
[![Release](https://github.com/filmil/webapp/actions/workflows/release.yml/badge.svg)](https://github.com/filmil/webapp/actions/workflows/release.yml)

## go-app.dev with Bazel

This is an interactive map application built using
[go-app.dev](https://go-app.dev/) and the [Bazel](https://bazel.build/) build
system with [rules_go](https://github.com/bazelbuild/rules_go).

## Features

- Interactive Map application using go-app.dev v9 and Leaflet.js.
- Address geocoding using Nominatim (OpenStreetMap).
- Reachable area (isochrone) computation using OpenRouteService API.
  Displays driving ranges at 1, 5, 10, 15, 20, 30, 40, and 60 minutes.
- Bazel configuration for both WASM and host server binaries.
- Integrated dependency management using Go modules and Bazel Bzlmod.
- Configured with `rules_shell` for shell testing.

## Requirements

To use the reachable area (isochrone) feature, you will need a free API key
from [OpenRouteService](https://openrouteservice.org/). The basic map view
and address search (geocoding) will work without an API key.

## Understanding Isochrones

An **isochrone** is a line drawn on a map connecting points at which something
occurs or arrives at the same time. In this application, the isochrones
represent the reachable area by driving a car from the searched address.

When you search for a location with a valid OpenRouteService API key, the app:

1. Geocodes your address into latitude and longitude coordinates.
2. Requests driving isochrones from OpenRouteService for 1, 5, 10, 15, 20, 30,
   40, and 60-minute intervals.
3. Renders these time boundaries as a series of color-coded polygons on the
   map, progressing from light yellow (1 minute) to dark red (60 minutes).

## Building the project

To build all targets, including the WASM binary and the server:

```bash
bazel build //...
```

## Running and Using the Server

To run the server:

```bash
bazel run //cmd/app:server
```

Once the server is running, follow these steps to use the application:

1. Open your web browser and navigate to `http://localhost:7000`.
2. In the "Enter an address..." field, type a location (e.g., "Paris").
3. (Optional) To view isochrones, paste your OpenRouteService API key into the
   "ORS API Key" field.
4. Click the **Search & Compute Area** button.
5. The map will automatically pan and zoom to the location. If an API key was
   provided, the color-coded reachable areas will be drawn around the pin.

## Directory Structure

- `cmd/app/`: Contains the main Go application code and Bazel build targets.
- `ai/`: Directory for AI-related scripts and tools.
- `MODULE.bazel`: Bazel module definition and dependency configuration.
- `go.mod`: Go module definition.
