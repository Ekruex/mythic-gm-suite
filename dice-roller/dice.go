package dice

import (
	"math/rand"
	"time"
)

// Create a single random source
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Dice represents a single type of die (e.g., d6, d20)
type Dice struct {
	Sides int
}

// Roll rolls the dice once and returns the result
func (d Dice) Roll() int {
	return rng.Intn(d.Sides) + 1 // Random number between 1 and Sides
}
