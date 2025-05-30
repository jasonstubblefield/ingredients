# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go library that extracts and tags recipe ingredients from any recipe on the internet. It parses HTML from recipe websites to extract structured ingredient information including amounts, measurements, and ingredient names.

## Build Commands

```bash
# Generate corpus data (required after modifying corpus source files)
go generate

# Build the CLI tool
go build -o ingredients cmd/ingredients/main.go

# Build the batch processing tool
go build -o ingredients-files cmd/ingredients-files/main.go

# Run all tests
go test -v ./...

# Run specific test
go test -v -run TestName

# Run benchmarks
go test -bench=. -benchmem

# Format code
gofmt -w .

# Build cross-platform releases
goreleaser release --snapshot --rm-dist
```

## Architecture

The codebase follows a data-driven approach where ingredient recognition is based on pre-compiled corpus data:

1. **Corpus Generation Flow**: JSON/text files in `corpus/` → `corpus/main.go` generator → `corpus.go` (auto-generated)
   - The corpus contains known ingredients, measurements, numbers, and conversion factors
   - Run `go generate` to rebuild corpus.go after modifying source data

2. **Ingredient Parsing Flow**: HTML → Recipe struct → sanitized text → tokenization → tag identification → normalized measurements
   - Uses the corpus data to identify and normalize ingredients
   - Converts all measurements to cups for standardization

3. **Command-Line Tools**:
   - `cmd/ingredients/` - Processes single files/URLs
   - `cmd/ingredients-files/` - Parallel batch processing with progress bars

## Key Implementation Details

- **Measurement Normalization**: All measurements are converted to cups using density data and conversion factors in the corpus
- **Text Sanitization**: Handles fractions, unicode, parentheses, and special formatting before parsing
- **Testing Strategy**: Uses real HTML files from recipe websites in `testing/sites/` as test fixtures
- **Performance**: Batch processing tool uses goroutines for parallel processing with configurable worker count