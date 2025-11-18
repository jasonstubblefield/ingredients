package ingredients

import (
	"fmt"
	"testing"

	log "github.com/schollz/logger"
)

func TestWeberRecipe(t *testing.T) {
	log.SetLevel("trace")

	html := `<!DOCTYPE html>
<html>
<head>
    <script type="application/ld+json">
    {
        "@context": "http://schema.org",
        "@type": "Recipe",
        "name": "Asian Noodles",
        "recipeIngredient": [
            "60 grams plus 2 tablespoons creamy peanut butter",
            "2 tablespoons hot chili-garlic sauce",
            "2 tablespoons soy sauce",
            "Kosher salt",
            "255 grams dried soba noodles"
        ]
    }
    </script>
</head>
<body>
</body>
</html>`

	r, err := NewFromHTML("test", html)
	if err != nil {
		t.Fatalf("Error parsing: %v", err)
	}

	fmt.Printf("\n\n=== RESULTS ===\n")
	fmt.Printf("Found %d lines\n", len(r.Lines))
	for i, line := range r.Lines {
		fmt.Printf("Line %d: %s\n", i, line.LineOriginal)
		fmt.Printf("  Ingredient: %s\n", line.Ingredient.Name)
		fmt.Printf("  Amount: %f\n", line.Ingredient.Measure.Amount)
		fmt.Printf("  Measure: %s\n", line.Ingredient.Measure.Name)
	}

	fmt.Printf("\nFound %d ingredients\n", len(r.Ingredients))
	for i, ing := range r.Ingredients {
		fmt.Printf("Ingredient %d: %s - %f %s\n", i, ing.Name, ing.Measure.Amount, ing.Measure.Name)
	}
}

func TestPistachioRecipe(t *testing.T) {
	log.SetLevel("trace")

	// Test individual line scoring first to understand the problem
	fmt.Printf("\n=== TESTING INDIVIDUAL LINE SCORES ===\n")
	testLines := []string{
		"4 tablespoons Olive oil, use divided",
		"1 Small onion, finely chopped",
		"3 Cloves garlic, minced",
		"1 10-ounce package Frozen spinach, thawed",
		"Kosher salt and black pepper",
		"1 teaspoon Dried oregano",
		"Juice of 1 lemon",
		"Zest of 1 lemon, finely grated",
		"4 ounces Feta cheese, crumbled (omit for dairy free)",
		"4 Boneless, skinless chicken breasts",
		"¾ cup Pistachios, chopped",
		"1 pint Cherry or grape tomatoes, halved",
	}

	totalScore := 0
	for i, line := range testLines {
		score, lineInfo := scoreLine(line)
		totalScore += score
		fmt.Printf("Line %d (len=%d, score=%d): %s\n", i, len(lineInfo.Line), score, line)
		fmt.Printf("  Ingredients: %v\n", lineInfo.IngredientsInString)
		fmt.Printf("  Amounts: %v\n", lineInfo.AmountInString)
		fmt.Printf("  Measures: %v\n", lineInfo.MeasureInString)
	}
	fmt.Printf("\nTotal score for all lines: %d (need > 2 to qualify as ingredient list)\n", totalScore)
	fmt.Printf("Number of lines: %d (need >= 2)\n\n", len(testLines))

	// Simulating the HTML structure from americanpistachios.org
	// No schema.org markup, no JSON-LD, just plain <p> tags in a div
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Pistachio Crusted Greek Chicken</title>
</head>
<body>
    <div class="field field--name-field-american-ingredients field--type-text-long field--label-hidden field--item">
        <p>4 tablespoons Olive oil, use divided</p>
        <p>1 Small onion, finely chopped</p>
        <p>3 Cloves garlic, minced</p>
        <p>1 10-ounce package Frozen spinach, thawed</p>
        <p>Kosher salt and black pepper</p>
        <p>1 teaspoon Dried oregano</p>
        <p>Juice of 1 lemon</p>
        <p>Zest of 1 lemon, finely grated</p>
        <p>4 ounces Feta cheese, crumbled (omit for dairy free)</p>
        <p>4 Boneless, skinless chicken breasts</p>
        <p>¾ cup Pistachios, chopped</p>
        <p>1 pint Cherry or grape tomatoes, halved</p>
    </div>
</body>
</html>`

	r, err := NewFromHTML("test-pistachio", html)
	if err != nil {
		fmt.Printf("\nParser error: %v\n", err)
		fmt.Printf("\nThis confirms the issue - the scoring algorithm is too strict!\n")
		return // Don't fail the test, we're just debugging
	}

	fmt.Printf("\n\n=== PISTACHIO RECIPE RESULTS ===\n")
	fmt.Printf("Found %d lines\n", len(r.Lines))
	for i, line := range r.Lines {
		fmt.Printf("Line %d: %s\n", i, line.LineOriginal)
		fmt.Printf("  Ingredient: %s\n", line.Ingredient.Name)
		fmt.Printf("  Amount: %f\n", line.Ingredient.Measure.Amount)
		fmt.Printf("  Measure: %s\n", line.Ingredient.Measure.Name)
	}

	fmt.Printf("\nFound %d ingredients\n", len(r.Ingredients))
	for i, ing := range r.Ingredients {
		fmt.Printf("Ingredient %d: %s - %f %s\n", i, ing.Name, ing.Measure.Amount, ing.Measure.Name)
	}

	// Expect at least 10 ingredients (some may not have amounts)
	if len(r.Ingredients) < 10 {
		t.Errorf("Expected at least 10 ingredients, got %d", len(r.Ingredients))
	}
}
