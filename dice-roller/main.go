package main

import (
	"fmt"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
	"github.com/Ekruex/mythic-gm-suite/dice-roller/roller"
)

func main() {
	// Standard d20 roll with a +5 modifier
	modifier := 5
	roll := dice.D20.Roll(modifier)
	fmt.Println("Standard d20 Roll (+5):", roll)

	// Fortune roll (Advantage) with a +3 modifier
	fortuneDetails, fortuneResult := roller.RollWithFortune(dice.D20, 3)
	fmt.Printf("Fortune Roll (+3): %s → Final: %d\n", fortuneDetails, fortuneResult)

	// Misfortune roll (Disadvantage) with a -2 modifier
	misfortuneDetails, misfortuneResult := roller.RollWithMisfortune(dice.D20, -2)
	fmt.Printf("Misfortune Roll (-2): %s → Final: %d\n", misfortuneDetails, misfortuneResult)

	// Rolling a d6 with no modifier
	d6Roll := dice.D6.Roll(0)
	fmt.Println("Standard d6 Roll:", d6Roll)
}
