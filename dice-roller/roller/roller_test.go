package roller

import (
	"testing"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
)

func TestRollMultiple(t *testing.T) {
	d := dice.NewDice(6)
	rolls := []dice.DiceRoll{d.Roll(), d.Roll(), d.Roll()}
	results, formattedResult, err := RollMultiple(rolls, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	total := Sum(results)
	expectedFormattedResult := FormatRollResult(results, 3)
	if formattedResult != expectedFormattedResult {
		t.Errorf("Expected formatted result %s, got %s", expectedFormattedResult, formattedResult)
	}
	if total < 3 || total > 18 {
		t.Errorf("Total out of expected range: %d", total)
	}
}

func TestRollWithFortune(t *testing.T) {
	d := dice.NewDice(20)
	details, result, err := RollWithFortune(d, 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result < 6 || result > 25 {
		t.Errorf("Result out of expected range: %d", result)
	}
	if details == "" {
		t.Errorf("Expected details to be non-empty")
	}
}

func TestRollWithMisfortune(t *testing.T) {
	d := dice.NewDice(20)
	details, result, err := RollWithMisfortune(d, -2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result < -1 || result > 18 {
		t.Errorf("Result out of expected range: %d", result)
	}
	if details == "" {
		t.Errorf("Expected details to be non-empty")
	}
}

func TestFormatRollResult(t *testing.T) {
	results := []int{6, 1}
	modifier := 12
	expected := "6 + 1 + 12 = 19"
	formattedResult := FormatRollResult(results, modifier)
	if formattedResult != expected {
		t.Errorf("Expected formatted result %s, got %s", expected, formattedResult)
	}
}

func TestSum(t *testing.T) {
	results := []int{6, 1, 12}
	expected := 19
	total := Sum(results)
	if total != expected {
		t.Errorf("Expected sum %d, got %d", expected, total)
	}
}
