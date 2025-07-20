# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LogInform is a Go CLI application that helps find and explain log identifiers in text content using predefined explanations. The application loads explanations from a YAML file and provides commands to either find matching identifiers in provided content or inspect all loaded explanations.

## Architecture

The codebase follows a simple and clean architecture:

- **main.go**: Entry point using Cobra CLI framework with two main commands (`explain` and `inspect`)
- **internal/explanation.go**: Core data structures and YAML loading functionality for `Explanation` structs
- **internal/find.go**: Content processing logic that searches for identifiers and prints matches with explanations
- **internal/inspect.go**: Utility for displaying all loaded explanations
- **explanations.yml**: Configuration file containing log identifiers with internet and internal explanations

The application loads explanations once at startup (via `PersistentPreRunE`) and uses them across all commands. Each explanation contains an identifier, internet explanation, and internal explanation.

## Development Commands

### Build and Run
```bash
go run .                    # Run the application
make run                    # Alternative run command
```

### Testing
```bash
go test ./...               # Run all tests
go test -race ./...         # Run tests with race detection
make check                  # Full test suite with formatting, security, and linting
```

### Code Quality
```bash
make check                  # Comprehensive check: format, security, static analysis, tests
goimports -w .              # Format and organize imports
gofmt -w .                  # Format code
go mod tidy                 # Clean up dependencies
govulncheck ./...           # Security vulnerability check
go vet ./...                # Static analysis
staticcheck ./...           # Additional static analysis
golangci-lint run --fix     # Lint and auto-fix
```

### Development Workflow
```bash
make poll                   # Watch Go files and run checks on changes (uses entr)
make doctor                 # Full validation including Docker build and run
```

### Docker
```bash
docker build -t loginform . # Build Docker image
docker run --rm loginform   # Run in container (runs tests)
```

## Application Usage

The application expects an `explanations.yml` file in the current directory and provides two commands:

- `explain <content>`: Find lines containing identifiers and show explanations
- `inspect`: Display all loaded explanations

## Dependencies

- **cobra**: CLI framework for command structure
- **testify**: Testing framework with assertions
- **yaml.v3**: YAML parsing for explanations file
- Uses Go 1.24.5 with standard library for most functionality