package roller

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
)

// RollLog stores previous roll results
var (
	rollLog    []string
	mu         sync.Mutex // Ensures safe concurrent access
	maxLogSize = 100      // Limit the roll history to 100 entries
)

func RollMultiple(d dice.Dice, times int, modifier ...int) ([]int, string, error) {
	if times <= 0 {
		return nil, "", errors.New("times must be greater than 0")
	}
	results := make([]int, times)
	total := 0
	mod := getModifier(modifier)

	for i := 0; i < times; i++ {
		results[i] = d.Roll(0) // Roll without modifier for individual dice
		total += results[i]
	}

	total += mod // Add modifier to the total

	// Check for critical success or failure
	criticalMessage := ""
	if d.Sides == 20 {
		for _, result := range results {
			if result == 20 {
				criticalMessage = "Critical Success!"
			} else if result == 1 {
				criticalMessage = "Critical Failure!"
			}
		}
	}

	// Format the result
	formattedResult := FormatRollResult(results, mod)
	if criticalMessage != "" {
		formattedResult += fmt.Sprintf(" (%s)", criticalMessage)
	}

	fmt.Printf("Formatted roll result: %s\n", formattedResult) // Debug log
	storeRoll(formattedResult)

	// Return the results, formatted result, and no error
	return results, formattedResult, nil
}

// Store roll result in history
func storeRoll(entry string) {
	mu.Lock()
	defer mu.Unlock()

	fmt.Printf("Storing roll: %s\n", entry) // Debug log

	// Add the new entry to the log
	rollLog = append(rollLog, entry)

	// Debug log to show the updated rollLog
	fmt.Printf("Updated rollLog before trimming: %v\n", rollLog)

	// Enforce the maximum log size
	if len(rollLog) > maxLogSize {
		rollLog = rollLog[1:]                                     // Remove the oldest entry
		fmt.Printf("Roll log trimmed to max size: %v\n", rollLog) // Debug log
	}

	// Final debug log to confirm the state of rollLog
	fmt.Printf("Final rollLog after trimming: %v\n", rollLog)
}

// GetRollHistory returns all previous roll results
func GetRollHistory() []string {
	mu.Lock()
	defer mu.Unlock()

	// Debug log before returning the roll history
	fmt.Printf("Retrieving roll history: %v\n", rollLog) // Debug log

	return append([]string{}, rollLog...) // Return a copy to prevent modification
}

// ClearRollHistory clears the roll history
func ClearRollHistory() {
	mu.Lock()
	defer mu.Unlock()

	fmt.Println("Clearing roll history...") // Debug log
	rollLog = []string{}
	fmt.Printf("Roll history after clearing: %v\n", rollLog) // Debug log
}

// RollWithFortune rolls two d20s, displays both, and returns the highest before applying modifiers
func RollWithFortune(d dice.Dice, modifier ...int) (string, int, error) {
	if d.Sides != 20 {
		roll := d.Roll(getModifier(modifier))
		storeRoll(fmt.Sprintf("%d", roll))
		return fmt.Sprintf("%d", roll), roll, nil // Fortune only applies to d20 rolls
	}
	r1, r2 := d.Roll(0), d.Roll(0)
	highest := int(math.Max(float64(r1), float64(r2)))
	mod := getModifier(modifier)
	total := highest + mod

	// Check for critical success or failure
	criticalMessage := ""
	if r1 == 20 || r2 == 20 {
		criticalMessage = "Critical Success!"
	} else if r1 == 1 || r2 == 1 {
		criticalMessage = "Critical Failure!"
	}

	details := fmt.Sprintf("%d / %d", r1, r2)
	finalResult := fmt.Sprintf("%s + %d = %d", details, mod, total)
	if criticalMessage != "" {
		finalResult += fmt.Sprintf(" (%s)", criticalMessage)
	}
	storeRoll(finalResult) // Store total result in history

	return finalResult, total, nil
}

// RollWithMisfortune rolls two d20s, displays both, and returns the lowest before applying modifiers
func RollWithMisfortune(d dice.Dice, modifier ...int) (string, int, error) {
	if d.Sides != 20 {
		roll := d.Roll(getModifier(modifier))
		storeRoll(fmt.Sprintf("%d", roll))
		return fmt.Sprintf("%d", roll), roll, nil // Misfortune only applies to d20 rolls
	}
	r1, r2 := d.Roll(0), d.Roll(0)
	lowest := int(math.Min(float64(r1), float64(r2)))
	mod := getModifier(modifier)
	total := lowest + mod

	// Check for critical success or failure
	criticalMessage := ""
	if r1 == 20 || r2 == 20 {
		criticalMessage = "Critical Success!"
	} else if r1 == 1 || r2 == 1 {
		criticalMessage = "Critical Failure!"
	}

	details := fmt.Sprintf("%d / %d", r1, r2)
	finalResult := fmt.Sprintf("%s + %d = %d", details, mod, total)
	if criticalMessage != "" {
		finalResult += fmt.Sprintf(" (%s)", criticalMessage)
	}
	storeRoll(finalResult) // Store total result in history

	return finalResult, total, nil
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
		if modifier > 0 {
			parts = append(parts, fmt.Sprintf("%d", modifier))
		} else {
			parts = append(parts, fmt.Sprintf("%d", -modifier))
		}
	}

	// Convert to final output string
	finalResult := fmt.Sprintf("%s = %d", strings.Join(parts, " + "), sum(results)+modifier)
	return finalResult
}

// Helper function to sum up results
func sum(results []int) int {
	total := 0
	for _, r := range results {
		total += r
	}
	return total
}
