package roller

import (
	"fmt"
	"math"
	"strings"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
)

// RollMultiple rolls a dice multiple times with an optional modifier
func RollMultiple(d dice.Dice, times int, modifier ...int) ([]int, int) {
	results := make([]int, times)
	total := 0
	mod := getModifier(modifier)

	for i := 0; i < times; i++ {
		results[i] = d.Roll(0) // Roll without modifier first
		total += results[i]
	}

	total += mod // Add modifier once at the end
	return results, total
}

// RollWithFortune rolls two d20s, displays both, and returns the highest before applying modifiers
func RollWithFortune(d dice.Dice, modifier ...int) (string, int) {
	if d.Sides != 20 {
		roll := d.Roll(getModifier(modifier))
		return fmt.Sprintf("%d", roll), roll // Fortune only applies to d20 rolls
	}
	r1, r2 := d.Roll(0), d.Roll(0)
	highest := int(math.Max(float64(r1), float64(r2)))
	mod := getModifier(modifier)

	return fmt.Sprintf("%d / %d", r1, r2), highest + mod
}

// RollWithMisfortune rolls two d20s, displays both, and returns the lowest before applying modifiers
func RollWithMisfortune(d dice.Dice, modifier ...int) (string, int) {
	if d.Sides != 20 {
		roll := d.Roll(getModifier(modifier))
		return fmt.Sprintf("%d", roll), roll // Misfortune only applies to d20 rolls
	}
	r1, r2 := d.Roll(0), d.Roll(0)
	lowest := int(math.Min(float64(r1), float64(r2)))
	mod := getModifier(modifier)

	return fmt.Sprintf("%d / %d", r1, r2), lowest + mod
}

// Helper function to safely retrieve the modifier
func getModifier(modifier []int) int {
	if len(modifier) > 0 {
		return modifier[0]
	}
	return 0
}

// FormatRollResult formats dice rolls like "7 + 14 + 5 + 3 + 2 = 31"
func FormatRollResult(results []int, modifier int) string {
	var parts []string

	// Convert each roll result to a string
	for _, roll := range results {
		parts = append(parts, fmt.Sprintf("%d", roll))
	}

	// Append the modifier separately
	if modifier != 0 {
		parts = append(parts, fmt.Sprintf("+ %d", modifier))
	}

	// Convert to final output string
	return fmt.Sprintf("%s = %d", strings.Join(parts, " + "), sum(results)+modifier)
}

// Helper function to sum up results
func sum(results []int) int {
	total := 0
	for _, r := range results {
		total += r
	}
	return total
}
