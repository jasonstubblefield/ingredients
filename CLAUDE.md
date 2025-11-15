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

# Run tests with debug logging
DEBUG=1 go test -v -run TestName

# Format code
gofmt -w .

# Build cross-platform releases
goreleaser release --snapshot --rm-dist
```

## Architecture

The codebase follows a data-driven approach where ingredient recognition is based on pre-compiled corpus data:

1. **Corpus Generation Flow**: JSON/text files in `corpus/` → `corpus/main.go` generator → `corpus.go` (auto-generated)
   - The corpus contains known ingredients (herbs, fruits, vegetables), measurements, numbers, conversion factors, and density data
   - Ingredients are sorted by length (longest first) then alphabetically for optimal matching
   - Run `go generate` to rebuild corpus.go after modifying source data

2. **Ingredient Parsing Flow**: HTML → Recipe struct → sanitized text → tokenization → tag identification → normalized measurements
   - HTML parsing attempts to extract JSON-LD structured data from `<script>` tags first (fast path)
   - Falls back to DOM traversal with scoring heuristics to identify ingredient sections
   - Scoring system evaluates lines based on: presence of ingredients, amounts, measurements, position ordering, length, and punctuation
   - Lines with score >2 in groups of 2-25 lines are considered ingredient lists
   - Uses the corpus data to identify and normalize ingredients via string matching
   - Converts all measurements to cups for standardization

3. **Command-Line Tools**:
   - `cmd/ingredients/` - Processes single files/URLs or stdin (for cached HTML)
   - `cmd/ingredients-files/` - Parallel batch processing with progress bars using all CPU cores

## Key Implementation Details

- **Corpus Data Sources**:
  - `corpus/ingredients.txt` - Main ingredient list
  - `corpus/herbs.json`, `corpus/fruits.json`, `corpus/vegetables.json` - Categorized ingredients
  - `corpus/densities.json` - Ingredient densities for volume/weight conversion
  - `corpus/numbers.txt` - Number words and fractions
  - `corpus/directions_pos.txt`, `corpus/directions_neg.txt` - Words for scoring

- **Measurement Normalization**: All measurements are converted to cups using density data and conversion factors in the corpus
  - Supported units: cup, tbl, tsp, ounce, gram, milliliter, pint, quart, pound, can
  - Fractions handled via unicode (½, ¼, ¾, etc.) and decimal conversion

- **Text Sanitization**: `SanitizeLine()` in utils.go handles fractions, unicode, parentheses, and special formatting before parsing

- **Testing Strategy**: Uses real HTML files from recipe websites in `testing/sites/` as test fixtures
  - Each site has subdirectories with recipe HTML and expected results
  - Set DEBUG=1 environment variable for trace-level logging during tests

- **Performance**: Batch processing tool uses goroutines for parallel processing across all CPU cores
  - Results cached as MD5-hashed JSON files to avoid reprocessing