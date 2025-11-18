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

## Critical Implementation Details

### UTF-8 and String Indexing
- **IMPORTANT**: The codebase uses a trie-based search (`utils.go:57`) that returns **rune positions**, not byte positions
- All position-based operations must account for multi-byte UTF-8 characters (like ½, ¾, ¼)
- `getOtherInBetweenPositions()` in `utils.go:285` correctly uses rune-based indexing to extract comments
- When debugging position issues, remember: `len(string)` returns bytes, `len([]rune(string))` returns characters

### Minimum Ingredient Count Checks
The parser has three critical locations that enforce minimum ingredient counts (currently >= 2):
1. `ingredients.go:396` - Schema.org extraction: `if schemaErr == nil && len(schemaLineInfos) >= 2`
2. `ingredients.go:416` - JSON extraction: `if errJSON == nil && len(lis) >= 2`
3. `ingredients.go:434` - DOM parsing: `if score > 2 && len(childrenLineInfo) < 25 && len(childrenLineInfo) >= 2`

### Scoring Algorithm for Ingredient Detection
- Located in `scoreLine()` function at `ingredients.go:533`
- Evaluates lines based on: ingredients present (+1), amounts present (+1), measures present (+1), correct ordering (+3)
- **Length penalty** at `ingredients.go:587`: Lines longer than 40 characters get penalized
  - This threshold was increased from 30 to 40 to handle verbose ingredient descriptions
  - Threshold too low = valid ingredients rejected; too high = non-ingredient text included
- Total score >2 required for a group of 2-25 lines to qualify as an ingredient list

### Comment Extraction Logic
- Comments are extracted from text between the measure and ingredient positions
- Process: `getOtherInBetweenPositions(lineInfo.Line, lineInfo.MeasureInString[0], lineInfo.IngredientsInString[0])`
- Example: "1 cup **cold** butter" → measure="cup" (pos 4), ingredient="butter" (pos 11), comment="cold"
- Only works when both measure and ingredient are detected in the line

### Parsing Pipeline Details
The `parseRecipe()` function (`ingredients.go:204`) filters and processes ingredient lines:
1. Filters out lines that are too short (<3) or too long (>150 for DOM, >250 for schema.org)
2. Filters out lines containing "serving size" or "yield"
3. Extracts amount, ingredient name, and measure for each line
4. Extracts comments between measure and ingredient positions
5. Normalizes all measurements to cups
6. Consolidates duplicate ingredients via `ConvertIngredients()`

### Test Expectations
- Test failures after parser improvements often indicate the tests were written for **buggy behavior**
- When fixing parser logic, expect to update test expectations to match the corrected output
- Use `update_tests.go` pattern to regenerate expectations from actual parser output
- Always verify fixes don't break existing working recipes (regression testing)