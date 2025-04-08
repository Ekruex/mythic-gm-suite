package dice

import (
	"errors" // For creating error messages
	"math/rand"
	"strconv" // For converting strings to integers
	"strings" // For string manipulation
	"time"
)

// Parse parses a dice notation string (e.g., "3d6+2d4+5") and returns a list of DiceRoll objects and a total modifier
func Parse(notation string) ([]DiceRoll, int, error) {
	// Split the notation into parts (e.g., "3d6+2d4+5" -> ["3d6", "2d4", "+5"])
	parts := strings.FieldsFunc(notation, func(r rune) bool {
		return r == '+' || r == '-'
	})

	var diceRolls []DiceRoll
	modifier := 0

	for _, part := range parts {
		if strings.Contains(part, "d") {
			// Parse dice notation (e.g., "3d6")
			diceParts := strings.Split(part, "d")
			if len(diceParts) != 2 {
				return nil, 0, errors.New("invalid dice notation")
			}

			// Parse the number of dice
			numDice, err := strconv.Atoi(diceParts[0])
			if err != nil || numDice <= 0 {
				numDice = 1 // Default to 1 die if not specified or invalid
			}

			// Parse the number of sides
			sides, err := strconv.Atoi(diceParts[1])
			if err != nil || sides <= 0 {
				return nil, 0, errors.New("invalid number of sides")
			}

			// Add the parsed dice roll to the list
			diceRolls = append(diceRolls, DiceRoll{
				Dice:     NewDice(sides),
				Count:    numDice,
				Modifier: 0,
			})
		} else {
			// Parse the modifier (e.g., "+5" or "-3")
			mod, err := strconv.Atoi(part)
			if err != nil {
				return nil, 0, errors.New("invalid modifier")
			}
			modifier += mod
		}
	}

	return diceRolls, modifier, nil
}

// DiceRoll represents a parsed dice roll
type DiceRoll struct {
	Dice     Dice
	Count    int
	Modifier int
}

// Dice represents a single die with a set number of sides
type Dice struct {
	Sides int
	rng   *rand.Rand
}

// NewDice creates a new Dice of a specified type
func NewDice(sides int) Dice {
	return Dice{
		Sides: sides,
		rng:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Roll rolls the dice and applies an optional modifier
func (d Dice) Roll(modifier ...int) int {
	mod := getModifier(modifier) // Get the first modifier or 0
	return d.rng.Intn(d.Sides) + 1 + mod
}

// Helper function to retrieve the modifier safely
func getModifier(modifier []int) int {
	if len(modifier) > 0 {
		return modifier[0]
	}
	return 0 // Default to 0 if no modifier is provided
}

// Predefined dice types
var (
	D4   = NewDice(4)
	D6   = NewDice(6)
	D8   = NewDice(8)
	D10  = NewDice(10)
	D12  = NewDice(12)
	D20  = NewDice(20)
	D100 = NewDice(100)
)
