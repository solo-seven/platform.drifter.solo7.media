package server_test

import (
	"testing"

	"github.com/solo-seven/platform.drifter.solo7.media/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExpressionExtractor_ExtractExpressions(t *testing.T) {
	mockLogger := &MockLogger{}
	extractor := server.NewExpressionExtractor(mockLogger)

	t.Run("extract expressions from data", func(t *testing.T) {
		data := map[string]interface{}{
			"effect": "heal(self, roll(\"1d10\") + self.level)",
			"damage": "roll(\"2d6\") + 3",
			"name":   "Basic Attack",
		}

		mockLogger.On("Debug", "Expressions extracted", mock.AnythingOfType("map[string]interface {}")).Return()

		expressions, err := extractor.ExtractExpressions(data)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		assert.Contains(t, expressions, "effect")
		assert.Contains(t, expressions, "damage")
		assert.Equal(t, "heal(self, roll(\"1d10\") + self.level)", expressions["effect"])
		assert.Equal(t, "roll(\"2d6\") + 3", expressions["damage"])

		mockLogger.AssertExpectations(t)
	})

	t.Run("extract expressions from nested data", func(t *testing.T) {
		data := map[string]interface{}{
			"abilities": map[string]interface{}{
				"second_wind": map[string]interface{}{
					"effect": "heal(self, roll(\"1d10\") + self.level)",
				},
			},
		}

		mockLogger.On("Debug", "Expressions extracted", mock.AnythingOfType("map[string]interface {}")).Return()

		expressions, err := extractor.ExtractExpressions(data)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		assert.Contains(t, expressions, "abilities.second_wind.effect")

		mockLogger.AssertExpectations(t)
	})

	t.Run("extract expressions from array data", func(t *testing.T) {
		data := map[string]interface{}{
			"effects": []interface{}{
				"heal(self, roll(\"1d10\") + self.level)",
				"deal(target, roll(\"2d6\") + 3)",
			},
		}

		mockLogger.On("Debug", "Expressions extracted", mock.AnythingOfType("map[string]interface {}")).Return()

		expressions, err := extractor.ExtractExpressions(data)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		assert.Contains(t, expressions, "effects[0]")
		assert.Contains(t, expressions, "effects[1]")

		mockLogger.AssertExpectations(t)
	})

	t.Run("no expressions found", func(t *testing.T) {
		data := map[string]interface{}{
			"name":        "Basic Attack",
			"description": "A simple attack",
			"damage":      "1d8",
		}

		mockLogger.On("Debug", "Expressions extracted", mock.AnythingOfType("map[string]interface {}")).Return()

		expressions, err := extractor.ExtractExpressions(data)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		// "1d8" should be detected as an expression due to dice notation
		assert.Contains(t, expressions, "damage")

		mockLogger.AssertExpectations(t)
	})

	t.Run("nil data", func(t *testing.T) {
		expressions, err := extractor.ExtractExpressions(nil)

		assert.NoError(t, err)
		assert.NotNil(t, expressions)
		assert.Empty(t, expressions)
	})
}

func TestExpressionExtractor_ValidateExpression(t *testing.T) {
	mockLogger := &MockLogger{}
	extractor := server.NewExpressionExtractor(mockLogger)

	t.Run("valid expression", func(t *testing.T) {
		expression := "heal(self, roll(\"1d10\") + self.level)"

		err := extractor.ValidateExpression(expression)

		assert.NoError(t, err)
	})

	t.Run("empty expression", func(t *testing.T) {
		err := extractor.ValidateExpression("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expression cannot be empty")
	})

	t.Run("unbalanced parentheses", func(t *testing.T) {
		expression := "heal(self, roll(\"1d10\") + self.level"

		err := extractor.ValidateExpression(expression)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unbalanced parentheses")
	})

	t.Run("unbalanced brackets", func(t *testing.T) {
		expression := "heal(self, roll(\"1d10\") + self.level["

		err := extractor.ValidateExpression(expression)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unbalanced")
	})

	t.Run("unbalanced braces", func(t *testing.T) {
		expression := "heal(self, roll(\"1d10\") + self.level{"

		err := extractor.ValidateExpression(expression)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unbalanced")
	})

	t.Run("invalid function name", func(t *testing.T) {
		expression := "invalid_function(self, roll(\"1d10\"))"

		err := extractor.ValidateExpression(expression)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid function name")
	})

	t.Run("invalid variable name", func(t *testing.T) {
		expression := "heal($123invalid, roll(\"1d10\"))"

		err := extractor.ValidateExpression(expression)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid variable name")
	})

	t.Run("invalid dice notation", func(t *testing.T) {
		expression := "heal(self, roll(\"0d6\"))"

		err := extractor.ValidateExpression(expression)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid dice notation")
	})

	t.Run("valid dice notation", func(t *testing.T) {
		expression := "heal(self, roll(\"1d6\"))"

		err := extractor.ValidateExpression(expression)

		assert.NoError(t, err)
	})

	t.Run("valid function names", func(t *testing.T) {
		validFunctions := []string{
			"roll(\"1d6\")",
			"heal(self, 10)",
			"deal(target, 5)",
			"has_tag(self, \"tag\")",
			"get_attribute(self, \"str\")",
			"set_attribute(self, \"str\", 15)",
			"add(5, 3)",
			"subtract(10, 3)",
			"multiply(2, 3)",
			"divide(10, 2)",
			"mod(10, 3)",
			"min(5, 10)",
			"max(5, 10)",
			"abs(-5)",
			"floor(3.7)",
			"ceil(3.2)",
			"round(3.5)",
			"random(1, 10)",
			"choose([1, 2, 3])",
		}

		for _, expr := range validFunctions {
			err := extractor.ValidateExpression(expr)
			assert.NoError(t, err, "Expression should be valid: %s", expr)
		}
	})

	t.Run("valid variable references", func(t *testing.T) {
		validVariables := []string{
			"$strength",
			"$constitution",
			"$dexterity",
			"$intelligence",
			"$wisdom",
			"$charisma",
			"$level",
			"$hit_points",
			"$armor_class",
		}

		for _, varName := range validVariables {
			expression := "get_attribute(self, " + varName + ")"
			err := extractor.ValidateExpression(expression)
			assert.NoError(t, err, "Variable should be valid: %s", varName)
		}
	})

	t.Run("reserved words as variables", func(t *testing.T) {
		reservedWords := []string{"self", "target", "party", "terrain", "context"}

		for _, reserved := range reservedWords {
			expression := "get_attribute(" + reserved + ", \"str\")"
			err := extractor.ValidateExpression(expression)
			assert.NoError(t, err, "Reserved word should be valid in context: %s", reserved)
		}
	})
}

func TestExpressionExtractor_ParseExpression(t *testing.T) {
	mockLogger := &MockLogger{}
	extractor := server.NewExpressionExtractor(mockLogger)

	t.Run("parse valid expression", func(t *testing.T) {
		expression := "heal(self, roll(\"1d10\") + self.level)"

		mockLogger.On("Debug", "Expression parsed", mock.AnythingOfType("map[string]interface {}")).Return()

		result, err := extractor.ParseExpression(expression)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		// Check that result is a map with expected fields
		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "expression", resultMap["type"])
		assert.Equal(t, expression, resultMap["value"])

		mockLogger.AssertExpectations(t)
	})

	t.Run("parse empty expression", func(t *testing.T) {
		result, err := extractor.ParseExpression("")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expression cannot be empty")
		assert.Nil(t, result)
	})

	t.Run("parse complex expression", func(t *testing.T) {
		expression := "heal(self, roll(\"1d10\") + self.level + $constitution_modifier)"

		mockLogger.On("Debug", "Expression parsed", mock.AnythingOfType("map[string]interface {}")).Return()

		result, err := extractor.ParseExpression(expression)

		assert.NoError(t, err)
		assert.NotNil(t, result)

		mockLogger.AssertExpectations(t)
	})
}
