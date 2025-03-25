package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
	"github.com/Ekruex/mythic-gm-suite/dice-roller/roller"
)

func main() {
	http.HandleFunc("/api/roll", handleRoll)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/clear-history", handleClearHistory)

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRoll(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		http.Error(w, "Missing roll prompt", http.StatusBadRequest)
		return
	}

	_, result, err := parseAndRoll(prompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Roll Result: %s", result)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	history := roller.GetRollHistory()
	for _, entry := range history {
		fmt.Fprintf(w, "%s\n", entry)
	}
}

func handleClearHistory(w http.ResponseWriter, r *http.Request) {
	roller.ClearRollHistory()
	fmt.Fprintln(w, "Roll history cleared")
}

func parseAndRoll(prompt string) ([]int, string, error) {
	re := regexp.MustCompile(`(\d*)d(\d+)([+-]\d+)?`)
	matches := re.FindStringSubmatch(prompt)
	if len(matches) == 0 {
		return nil, "", fmt.Errorf("Invalid roll prompt")
	}

	numDice, _ := strconv.Atoi(matches[1])
	if numDice == 0 {
		numDice = 1
	}
	diceSides, _ := strconv.Atoi(matches[2])
	modifier := 0
	if matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	d := dice.NewDice(diceSides)
	results, formattedResult, err := roller.RollMultiple(d, numDice, modifier)
	if err != nil {
		return nil, "", err
	}

	return results, formattedResult, nil
}
