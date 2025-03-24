package main

import (
	"fmt"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
	"github.com/Ekruex/mythic-gm-suite/dice-roller/roller"
)

func main() {
	d := dice.Dice{Sides: 6}
	rolls := roller.RollMultiple(d, 3)
	fmt.Println(roller.FormatRollResult(d, 3, rolls))

	d20 := dice.Dice{Sides: 20}
	rolls = roller.RollMultiple(d20, 1)
	fmt.Println(roller.FormatRollResult(d20, 1, rolls))
}
