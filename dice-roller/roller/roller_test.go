package roller

import (
	"testing"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
)

func TestRollMultiple(t *testing.T) {
	d := dice.NewDice(6)
	results, total, err := RollMultiple(d, 3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
	if total < 3 || total > 18 {
		t.Errorf("Total out of expected range: %d", total)
	}
}

func TestRollWithFortune(t *testing.T) {
	d := dice.NewDice(20)
	_, result, err := RollWithFortune(d, 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result < 6 || result > 25 {
		t.Errorf("Result out of expected range: %d", result)
	}
}

func TestRollWithMisfortune(t *testing.T) {
	d := dice.NewDice(20)
	_, result, err := RollWithMisfortune(d, -2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result < -1 || result > 18 {
		t.Errorf("Result out of expected range: %d", result)
	}
}
