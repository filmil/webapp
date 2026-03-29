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

## Building the project

To build all targets, including the WASM binary and the server:

```bash
bazel build //...
```

## Running the application

To run the server:

```bash
bazel run //cmd/app:server
```

The application will be available at `http://localhost:7000`.

## Directory Structure

- `cmd/app/`: Contains the main Go application code and Bazel build targets.
- `ai/`: Directory for AI-related scripts and tools.
- `MODULE.bazel`: Bazel module definition and dependency configuration.
- `go.mod`: Go module definition.
