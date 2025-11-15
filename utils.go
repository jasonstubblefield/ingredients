package ingredients

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/schollz/logger"
)

var (
	reParentheses  = regexp.MustCompile(`(?s)\((.*)\)`)
	reNonAlphaNum  = regexp.MustCompile("[^a-zA-Z0-9/.]+")
	reNumberAtStart = regexp.MustCompile(`^\s*(\d+(?:\.\d+)?|\d+\s+\d+/\d+)`)
)

// Trie node for efficient pattern matching
type trieNode struct {
	children map[rune]*trieNode
	isEnd    bool
	value    string
}

// Trie for fast multi-pattern matching
type Trie struct {
	root *trieNode
}

// newTrie creates a new trie from a list of patterns
func newTrie(patterns []string) *Trie {
	t := &Trie{root: &trieNode{children: make(map[rune]*trieNode)}}
	for _, pattern := range patterns {
		t.insert(pattern)
	}
	return t
}

// insert adds a pattern to the trie
func (t *Trie) insert(pattern string) {
	node := t.root
	for _, ch := range pattern {
		if node.children[ch] == nil {
			node.children[ch] = &trieNode{children: make(map[rune]*trieNode)}
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.value = pattern
}

// findAll returns all pattern matches in the text with their positions
// Implements greedy longest-match strategy like the original algorithm
func (t *Trie) findAll(text string) []WordPosition {
	var matches []WordPosition
	runes := []rune(text)
	used := make([]bool, len(runes)) // Track which positions are already matched

	// Try to find matches starting at each position
	for i := 0; i < len(runes); i++ {
		if used[i] {
			continue
		}

		// Find longest match starting at position i
		node := t.root
		longestMatch := ""
		longestEnd := -1

		for j := i; j < len(runes); j++ {
			ch := runes[j]
			if node.children[ch] == nil {
				break
			}
			node = node.children[ch]
			if node.isEnd {
				longestMatch = node.value
				longestEnd = j
			}
		}

		// If we found a match, record it and mark positions as used
		if longestMatch != "" {
			matches = append(matches, WordPosition{
				Word:     strings.TrimSpace(longestMatch),
				Position: i,
			})
			// Mark these positions as used
			for j := i; j <= longestEnd; j++ {
				used[j] = true
			}
		}
	}

	// Sort by position
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Position < matches[j].Position
	})

	return matches
}

var (
	ingredientsTrie *Trie
	measuresTrie    *Trie
	numbersTrie     *Trie
)

var wordNumbers = map[string]float64{
	"zero": 0, "one": 1, "two": 2, "three": 3, "four": 4,
	"five": 5, "six": 6, "seven": 7, "eight": 8, "nine": 9,
	"ten": 10, "eleven": 11, "twelve": 12, "thirteen": 13,
	"fourteen": 14, "fifteen": 15, "sixteen": 16, "seventeen": 17,
	"eighteen": 18, "nineteen": 19, "twenty": 20, "thirty": 30,
	"forty": 40, "fifty": 50, "sixty": 60, "seventy": 70,
	"eighty": 80, "ninety": 90, "hundred": 100,
	"a": 1, "an": 1, "couple": 2, "few": 3, "several": 3,
	"half": 0.5, "quarter": 0.25,
}

// init builds the tries from corpus data for fast pattern matching
func init() {
	ingredientsTrie = newTrie(corpusIngredients)
	measuresTrie = newTrie(corpusMeasures)
	numbersTrie = newTrie(corpusNumbers)
}

// ConvertStringToNumber converts string numbers (including fractions and word forms) to float64
func ConvertStringToNumber(s string) float64 {
	// Handle unicode fractions
	switch s {
	case "½":
		return 0.5
	case "¼":
		return 0.25
	case "¾":
		return 0.75
	case "⅛":
		return 1.0 / 8
	case "⅜":
		return 3.0 / 8
	case "⅝":
		return 5.0 / 8
	case "⅞":
		return 7.0 / 8
	case "⅔":
		return 2.0 / 3
	case "⅓":
		return 1.0 / 3
	}

	// Handle word numbers
	if val, ok := wordNumbers[strings.ToLower(strings.TrimSpace(s))]; ok {
		return val
	}

	// Try to parse as float
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

func AmountToString(amount float64) string {
	r, _ := parseDecimal(fmt.Sprintf("%2.10f", amount))
	rationalFraction := float64(r.n) / float64(r.d)
	if rationalFraction > 0 {
		bestFractionDiff := 1e9
		bestFraction := 0.0
		var fractions = map[float64]string{
			0:       "",
			1:       "",
			1.0 / 2: "1/2",
			1.0 / 3: "1/3",
			2.0 / 3: "2/3",
			1.0 / 6: "1/6",
			1.0 / 8: "1/8",
			3.0 / 8: "3/8",
			5.0 / 8: "5/8",
			7.0 / 8: "7/8",
			1.0 / 4: "1/4",
			3.0 / 4: "3/4",
		}
		for f := range fractions {
			currentDiff := math.Abs(f - rationalFraction)
			if currentDiff < bestFractionDiff {
				bestFraction = f
				bestFractionDiff = currentDiff
			}
		}
		if fractions[bestFraction] == "" {
			return strconv.FormatInt(int64(math.Round(amount)), 10)
		}
		if r.i > 0 {
			return strconv.FormatInt(r.i, 10) + " " + fractions[bestFraction]
		} else {
			return fractions[bestFraction]
		}
	}
	return strconv.FormatInt(r.i, 10)
}

// A rational number r is expressed as the fraction p/q of two integers:
// r = p/q = (d*i+n)/d.
type rational struct {
	i int64 // integer
	n int64 // fraction numerator
	d int64 // fraction denominator
}

func gcd(x, y int64) int64 {
	for y != 0 {
		x, y = y, x%y
	}
	return x
}

func parseDecimal(s string) (r rational, err error) {
	sign := int64(1)
	if strings.HasPrefix(s, "-") {
		sign = -1
	}
	p := strings.IndexByte(s, '.')
	if p < 0 {
		p = len(s)
	}
	if i := s[:p]; len(i) > 0 {
		if i != "+" && i != "-" {
			r.i, err = strconv.ParseInt(i, 10, 64)
			if err != nil {
				return rational{}, err
			}
		}
	}
	if p >= len(s) {
		p = len(s) - 1
	}
	if f := s[p+1:]; len(f) > 0 {
		n, err := strconv.ParseUint(f, 10, 64)
		if err != nil {
			return rational{}, err
		}
		d := math.Pow10(len(f))
		if math.Log2(d) > 63 {
			err = fmt.Errorf(
				"ParseDecimal: parsing %q: value out of range", f,
			)
			return rational{}, err
		}
		r.n = int64(n)
		r.d = int64(d)
		if g := gcd(r.n, r.d); g != 0 {
			r.n /= g
			r.d /= g
		}
		r.n *= sign
	}
	return r, nil
}

// GetIngredientsInString returns the word positions of the ingredients
func GetIngredientsInString(s string) (wordPositions []WordPosition) {
	return ingredientsTrie.findAll(s)
}

// GetNumbersInString returns the word positions of the numbers in the ingredient string
func GetNumbersInString(s string) (wordPositions []WordPosition) {
	return numbersTrie.findAll(s)
}

// GetMeasuresInString returns the word positions of the measures in a ingredient string
func GetMeasuresInString(s string) (wordPositions []WordPosition) {
	return measuresTrie.findAll(s)
}

// WordPosition shows a word and its position
// Note: the position is memory-dependent as it will
// be the position after the last deleted word
type WordPosition struct {
	Word     string
	Position int
}

// getOtherInBetweenPositions returns the word positions comment string in the ingredients
func getOtherInBetweenPositions(s string, pos1, pos2 WordPosition) (other string) {
	if pos1.Position > pos2.Position {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Error(s, pos1, pos2)
			log.Error(r)
		}
	}()
	other = s[pos1.Position+len(pos1.Word)+1 : pos2.Position]
	other = strings.TrimSpace(other)
	return
}

// Replacer for common substitutions in SanitizeLine
var sanitizeReplacer = strings.NewReplacer(
	"⁄", "/",
	" / ", "/",
	"butter milk", "buttermilk",
	"bicarbonate of soda", "baking soda",
	"soda bicarbonate", "baking soda",
	" one ", " 1 ",
)

// SanitizeLine removes parentheses, trims the line, converts to lower case,
// replaces fractions with unicode and then does special conversion for ingredients (like eggs).
func SanitizeLine(s string) string {
	s = strings.ToLower(s)

	// Apply multiple replacements in one pass
	s = sanitizeReplacer.Replace(s)

	// remove parentheses
	for _, m := range reParentheses.FindAllStringSubmatch(s, -1) {
		s = strings.Replace(s, m[0], " ", 1)
	}

	s = " " + strings.TrimSpace(s) + " "

	// replace unicode fractions with fractions
	for v := range corpusFractionNumberMap {
		s = strings.Replace(s, v, " "+corpusFractionNumberMap[v].fractionString+" ", -1)
	}

	// remove non-alphanumeric
	s = reNonAlphaNum.ReplaceAllString(s, " ")

	// replace fractions with unicode fractions
	for v := range corpusFractionNumberMap {
		s = strings.Replace(s, corpusFractionNumberMap[v].fractionString, " "+v+" ", -1)
	}

	return s
}

var gramConversions = map[string]float64{
	"ounce": 28.3495,
	"gram":  1,
	"pound": 453.592,
}

var conversionToCup = map[string]float64{
	"tbl":        0.0625,
	"tsp":        0.020833,
	"cup":        1.0,
	"pint":       2.0,
	"quart":      4.0,
	"gallon":     16.0,
	"milliliter": 0.00423,
	"can":        1.75,
}
var ingredientToCups = map[string]float64{
	"eggs":      0.125,
	"egg":       0.125,
	"limes":     0.125,
	"lime":      0.125,
	"lemons":    0.1875,
	"lemon":     0.1875,
	"orange":    0.281,
	"oranges":   0.281,
	"egg yolks": 0.0625,
	"egg yolk":  0.0625,
	"garlic":    0.0280833,
	"chicken":   3,
	"celery":    0.5,
	"onion":     1,
	"carrot":    1,
	"butter":    0.5,
}

func cupsToOther(cups float64, ingredient string) (amount float64, measure string) {
	if _, ok := ingredientToCups[ingredient]; ok {
		measure = "whole"
		amount = math.Round(cups / ingredientToCups[ingredient])
		return
	}
	if cups > 0.125 {
		amount = cups
		measure = "cup"
	} else if cups > 0.020833*3 {
		amount = cups * 16
		measure = "tbl"
	} else {
		amount = cups * 48
		measure = "tsp"
	}
	if math.IsInf(amount, 0) {
		amount = 0
	}

	return
}

// normalizeIngredient will try to normalize the ingredient to 1 cup
func normalizeIngredient(ingredient, measure string, amount float64) (cups float64, err error) {
	// convert measure to standard measure
	newMeasure, ok := corpusMeasuresMap[measure]
	if !ok && measure != "whole" {
		err = fmt.Errorf("could not find '%s'", measure)
		return
	}
	measure = newMeasure
	if _, ok := ingredientToCups[ingredient]; ok && measure == "" {
		// special ingredients
		cups = amount * ingredientToCups[ingredient]
	} else if _, ok := conversionToCup[measure]; ok {
		// check if it has a standard volume measurement
		cups = float64(amount) * conversionToCup[measure]
	} else if _, ok := gramConversions[measure]; ok {
		// check if it has a standard weight measurement
		var density float64
		density, ok = densities[ingredient]
		if !ok {
			density = 200 // grams / cup
		}
		cups = amount * gramConversions[measure] / density
	} else {
		if _, ok := fruitMap[ingredient]; ok {
			cups = 1 * amount
		} else if _, ok := vegetableMap[ingredient]; ok {
			cups = 1 * amount
		} else if _, ok := herbMap[ingredient]; ok {
			cups = 0.0208333 * amount
		} else {
			err = errors.New("could not convert weight or volume")
		}
	}
	return
}

func determineMeasurementsFromCups(cups float64) (amount float64, measure string, amountString string, err error) {
	if cups > 0.125 {
		amount = cups
		measure = "cup"
	} else if cups > 0.020833*3 {
		amount = cups * 16
		measure = "tablespoon"
	} else {
		amount = cups * 48
		measure = "teaspoon"
	}
	amountString = AmountToString(amount)
	if math.IsInf(amount, 0) {
		amount = 0
	}
	if math.IsInf(cups, 0) {
		cups = 0
	}
	return
}
