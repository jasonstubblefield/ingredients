package ingredients

import (
	"testing"
	
	"github.com/stretchr/testify/assert"
)

func TestInternationalIngredients(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2 tablespoons gochujang", "gochujang"},
		{"1 cup tahini", "tahini"},
		{"3 sheets nori", "nori"},
		{"2 tbsp miso paste", "miso paste"},
		{"1/4 cup fish sauce", "fish sauce"},
		{"100g panko breadcrumbs", "panko breadcrumbs"},
		{"1 tsp za'atar", "zaatar"}, // Should match even with apostrophe
		{"2 cups plant milk", "plant milk"},
		{"1 can full fat coconut milk", "full fat coconut milk"},
	}
	
	for _, test := range tests {
		result, err := NewFromString(test.input)
		assert.NoError(t, err)
		ingredients := result.IngredientList()
		if len(ingredients.Ingredients) > 0 {
			assert.Equal(t, test.expected, ingredients.Ingredients[0].Name, 
				"Failed to parse: %s", test.input)
		}
	}
}

func TestInternationalDensities(t *testing.T) {
	// Test that new densities work for conversions
	tests := []struct {
		input       string
		expectedCups float64
	}{
		{"60g panko breadcrumbs", 1.0}, // 60g/60g per cup = 1 cup
		{"128g tahini", 0.5},           // 128g/256g per cup = 0.5 cups
	}
	
	for _, test := range tests {
		result, err := NewFromString(test.input)
		assert.NoError(t, err)
		ingredients := result.IngredientList()
		if len(ingredients.Ingredients) > 0 {
			assert.InDelta(t, test.expectedCups, 
				ingredients.Ingredients[0].Measure.Cups, 0.1,
				"Failed conversion for: %s", test.input)
		}
	}
}