package dice

import (
	"math/rand"
	"time"
)

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
