package roller

import (
	"fmt"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
)

// RollMultiple rolls multiple dice and returns the results
func RollMultiple(dice dice.Dice, times int) []int {
	results := make([]int, times)
	for i := 0; i < times; i++ {
		results[i] = dice.Roll()
	}
	return results
}

// FormatRollResult formats dice rolls nicely
func FormatRollResult(dice dice.Dice, times int, results []int) string {
	return fmt.Sprintf("%dd%d: %v", times, dice.Sides, results)
}
