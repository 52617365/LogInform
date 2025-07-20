# LogInform

A Go CLI application that helps find and explain log identifiers in text content using predefined explanations.

## Overview

LogInform searches through text content (files or inline strings) to identify specific log identifiers and provides both internet-sourced and internal explanations for each match. The application loads explanations from a YAML configuration file and offers commands to either find matching identifiers in provided content or inspect all loaded explanations.

## Features

- **Identifier Detection**: Searches text content for predefined log identifiers
- **Dual Explanations**: Provides both internet and internal explanations for each identifier
- **Flexible Input**: Accepts file paths or inline content via command-line flags
- **Rich Output**: Highlights matches with color coding and shows context
- **Configuration**: Uses YAML file for easy management of identifiers and explanations

## Installation

### Prerequisites

- Go 1.24.5 or later

### Build from Source

```bash
git clone https://github.com/52617365/LogInform.git
cd LogInform
go build -o loginform .
```

### Using Docker

```bash
docker build -t loginform .
docker run --rm -v $(pwd):/app loginform
```

## Usage

### Basic Commands

LogInform provides two main commands:

#### 1. Explain Command

Find and explain identifiers in content:

```bash
# Analyze a file
./loginform explain /path/to/logfile.txt

# Analyze inline content
./loginform explain --inline "ERROR_001 occurred in system"
```

#### 2. Inspect Command

View all loaded explanations:

```bash
./loginform inspect
```

### Configuration

Create an `explanations.yml` file in the same directory as the executable:

```yaml
- identifier: "ERROR_001"
  internetExplanation: "Network timeout error"
  internalExplanation: "Connection failed after 30 seconds"
- identifier: "WARN_002"
  internetExplanation: "Memory usage warning"
  internalExplanation: "Heap usage above 80% threshold"
- identifier: "INFO_003"
  internetExplanation: "System information message"
  internalExplanation: "Normal operation status update"
```

### Example Output

When running `./loginform explain --inline "System ERROR_001 detected"`:

```
Line 1:8-16 -> ERROR_001
  System ERROR_001 detected
  -> Internet: Network timeout error
  -> Internal: Connection failed after 30 seconds
```

## Development

### Project Structure

```
LogInform/
├── main.go                 # CLI entry point with Cobra commands
├── internal/
│   ├── explanation.go      # Core data structures and YAML loading
│   ├── explanation_test.go # Tests for explanation functionality
│   ├── inspect.go          # Inspection command implementation
│   └── inspect_test.go     # Tests for inspect functionality
├── explanations.yml        # Configuration file with identifiers
├── testing/
│   └── example_explain.txt # Example content for testing
├── Dockerfile             # Container build configuration
├── Makefile              # Development commands
└── CLAUDE.md             # Development guidance
```

### Development Commands

#### Build and Run
```bash
go run .                    # Run the application
make run                    # Alternative run command
```

#### Testing
```bash
go test ./...               # Run all tests
go test -race ./...         # Run tests with race detection
make check                  # Full test suite with formatting, security, and linting
```

#### Code Quality
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

#### Development Workflow
```bash
make poll                   # Watch Go files and run checks on changes (uses entr)
make doctor                 # Full validation including Docker build and run
```

### Architecture

The application follows a clean, modular architecture:

- **main.go**: Entry point using Cobra CLI framework with command definitions
- **internal/explanation.go**: Core functionality for loading and processing explanations
- **internal/inspect.go**: Utility for displaying all loaded explanations
- **explanations.yml**: YAML configuration file containing log identifiers and explanations

The application loads explanations once at startup and reuses them across all commands for optimal performance.

## Dependencies

- **[Cobra](https://github.com/spf13/cobra)**: CLI framework for command structure
- **[testify](https://github.com/stretchr/testify)**: Testing framework with assertions
- **[yaml.v3](https://github.com/go-yaml/yaml)**: YAML parsing for explanations file

## Contributing

1. Ensure Go 1.24.5+ is installed
2. Run `make check` before submitting changes
3. Add tests for new functionality
4. Update documentation as needed

## License

Apache 2.0
