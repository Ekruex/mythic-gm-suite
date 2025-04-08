package roller

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
)

var broadcast chan string // Declare the broadcast channel in the roller package

// SetBroadcastChannel sets the broadcast channel for the roller package
func SetBroadcastChannel(bc chan string) {
	broadcast = bc
}

// RollLog stores previous roll results
var (
	rollLog    []string
	mu         sync.Mutex // Ensures safe concurrent access
	maxLogSize = 10       // Limit the roll history to 10 entries
)

// HTTP handler for fetching roll history
func HandleFetchHistory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	history := GetRollHistory()
	json.NewEncoder(w).Encode(map[string]string{
		"history": strings.Join(history, "\n"),
	})
}

// HTTP handler for rolling dice
func HandleRoll(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Prompt   string `json:"prompt"`
		RollType string `json:"rollType"`
	}

	// Parse the JSON request
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Parse the dice prompt (e.g., "3d6+2d4+5")
	diceRolls, modifier, err := dice.Parse(request.Prompt)
	if err != nil {
		http.Error(w, "Invalid dice prompt", http.StatusBadRequest)
		return
	}

	// Perform the roll based on the roll type
	var results []int
	var formattedResult string
	var total int

	switch request.RollType {
	case "fortune":
		// Ensure only a single d20 roll is allowed for fortune
		if len(diceRolls) != 1 || diceRolls[0].Dice.Sides != 20 {
			http.Error(w, "Fortune rolls only apply to a single d20", http.StatusBadRequest)
			return
		}
		results, formattedResult, err = ParseAndRollWithFortune(request.Prompt)
		if err == nil {
			total = sum(results) + modifier
		}
	case "misfortune":
		// Ensure only a single d20 roll is allowed for misfortune
		if len(diceRolls) != 1 || diceRolls[0].Dice.Sides != 20 {
			http.Error(w, "Misfortune rolls only apply to a single d20", http.StatusBadRequest)
			return
		}
		results, formattedResult, err = ParseAndRollWithMisfortune(request.Prompt)
		if err == nil {
			total = sum(results) + modifier
		}
	default:
		// Roll all dice normally
		results, formattedResult, err = RollMultiple(diceRolls, modifier)
		if err == nil {
			total = sum(results) + modifier
		}
	}

	if err != nil {
		http.Error(w, "Error performing roll: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the roll result as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
		"total":   total,
		"result":  formattedResult,
	})
}

// ParseAndRollWithFortune parses the dice notation, rolls two dice, and returns the highest result
func ParseAndRollWithFortune(prompt string) ([]int, string, error) {
	// Parse the dice notation
	diceRoll, modifier, err := dice.Parse(prompt)
	if err != nil {
		return nil, "", err
	}

	// Ensure this is a d20 roll (Fortune only applies to d20)
	if len(diceRoll) != 1 || diceRoll[0].Dice.Sides != 20 {
		return nil, "", fmt.Errorf("fortune only applies to d20 rolls")
	}

	// Roll twice and take the higher result
	firstRoll := diceRoll[0].Dice.Roll()
	secondRoll := diceRoll[0].Dice.Roll()
	highestRoll := int(math.Max(float64(firstRoll), float64(secondRoll)))

	// Add the modifier to the highest roll
	total := highestRoll + modifier

	// Format the result
	formattedResult := fmt.Sprintf("%d / %d (Fortune: %d) + %d = %d", firstRoll, secondRoll, highestRoll, modifier, total)

	return []int{firstRoll, secondRoll}, formattedResult, nil
}

// ParseAndRollWithMisfortune parses the dice notation, rolls two dice, and returns the lowest result
func ParseAndRollWithMisfortune(prompt string) ([]int, string, error) {
	// Parse the dice notation
	diceRoll, modifier, err := dice.Parse(prompt)
	if err != nil {
		return nil, "", err
	}

	// Ensure this is a d20 roll (Misfortune only applies to d20)
	if len(diceRoll) != 1 || diceRoll[0].Dice.Sides != 20 {
		return nil, "", fmt.Errorf("misfortune only applies to d20 rolls")
	}

	// Roll twice and take the lower result
	firstRoll := diceRoll[0].Dice.Roll()
	secondRoll := diceRoll[0].Dice.Roll()
	lowestRoll := int(math.Min(float64(firstRoll), float64(secondRoll)))

	// Add the modifier to the lowest roll
	total := lowestRoll + modifier

	// Format the result
	formattedResult := fmt.Sprintf("%d / %d (Misfortune: %d) + %d = %d", firstRoll, secondRoll, lowestRoll, modifier, total)

	return []int{firstRoll, secondRoll}, formattedResult, nil
}

// ParseAndRoll parses the dice notation, rolls the specified number of dice, and returns the total result
func ParseAndRoll(prompt string) ([]int, string, error) {
	// Parse the dice notation
	diceRoll, modifier, err := dice.Parse(prompt)
	if err != nil {
		return nil, "", err
	}

	// Roll the dice
	results, formattedResult, err := RollMultiple(diceRoll, modifier)
	if err != nil {
		return nil, "", err
	}

	return results, formattedResult, nil
}

// RollMultiple rolls multiple dice types and returns the results, formatted result, and total
func RollMultiple(diceRolls []dice.DiceRoll, modifier int) ([]int, string, error) {
	var results []int
	total := 0
	var parts []string

	for _, dr := range diceRolls {
		for i := 0; i < dr.Count; i++ {
			roll := dr.Dice.Roll(0) // Roll the die
			results = append(results, roll)
			total += roll
			parts = append(parts, fmt.Sprintf("%d", roll))
		}
	}

	// Add the modifier to the total
	total += modifier

	// Append the modifier to the formatted result
	if modifier != 0 {
		parts = append(parts, fmt.Sprintf("%+d", modifier))
	}

	// Format the result
	formattedResult := fmt.Sprintf("%s = %d", strings.Join(parts, " + "), total)

	// Store the roll result in history
	storeRoll(formattedResult)

	return results, formattedResult, nil
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
		parts = append(parts, fmt.Sprintf("%+d", modifier))
	}

	// Convert to final output string
	finalResult := fmt.Sprintf("%s = %d", strings.Join(parts, " + "), sum(results)+modifier)
	return finalResult
}

// Store roll result in history
func storeRoll(entry string) {
	mu.Lock()
	defer mu.Unlock()

	// Add the new entry to the log
	rollLog = append(rollLog, entry)

	// Enforce the maximum log size
	if len(rollLog) > maxLogSize {
		rollLog = rollLog[1:] // Remove the oldest entry
	}

	// Reverse the rollLog for broadcasting
	reversedLog := make([]string, len(rollLog))
	for i, v := range rollLog {
		reversedLog[len(rollLog)-1-i] = v
	}

	// Broadcast the updated history
	updatedHistory := strings.Join(reversedLog, "\n")
	broadcast <- fmt.Sprintf(`{"type": "history", "history": %q}`, updatedHistory)

	// Debug log
	fmt.Printf("Broadcasting updated history: %s\n", updatedHistory)
}

// GetRollHistory returns all previous roll results
func GetRollHistory() []string {
	mu.Lock()
	defer mu.Unlock()
	return append([]string{}, rollLog...) // Return a copy to prevent modification
}

// ClearRollHistory clears the roll history
func ClearRollHistory() {
	mu.Lock()
	defer mu.Unlock()
	rollLog = []string{}
}

// Helper function to sum up results
func sum(results []int) int {
	total := 0
	for _, r := range results {
		total += r
	}
	return total
}
