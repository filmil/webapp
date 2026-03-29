# Webapp

This project was created with the help of Gemini CLI.

## Status

[![CI](https://github.com/filmil/webapp/actions/workflows/ci.yml/badge.svg)](https://github.com/filmil/webapp/actions/workflows/ci.yml)
[![Release](https://github.com/filmil/webapp/actions/workflows/release.yml/badge.svg)](https://github.com/filmil/webapp/actions/workflows/release.yml)

## go-app.dev with Bazel

This is a "Hello World" application built using [go-app.dev](https://go-app.dev/)
and the [Bazel](https://bazel.build/) build system with
[rules_go](https://github.com/bazelbuild/rules_go).

## Features

- Simple PWA "Hello World" using go-app.dev v9.
- Bazel configuration for both WASM and host server binaries.
- Integrated dependency management using Go modules and Bazel Bzlmod.

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
