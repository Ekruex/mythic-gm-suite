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
	http.HandleFunc("/api/fortune", handleFortune)
	http.HandleFunc("/api/misfortune", handleMisfortune)
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

func handleFortune(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		http.Error(w, "Missing roll prompt", http.StatusBadRequest)
		return
	}

	_, result, err := parseAndRollWithFortune(prompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Roll Result: %s", result)
}

func handleMisfortune(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		http.Error(w, "Missing roll prompt", http.StatusBadRequest)
		return
	}

	_, result, err := parseAndRollWithMisfortune(prompt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Roll Result: %s", result)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	history := roller.GetRollHistory()
	// Reverse the order of the history entries
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}
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
		return nil, "", fmt.Errorf("invalid roll prompt")
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

func parseAndRollWithFortune(prompt string) ([]int, string, error) {
	re := regexp.MustCompile(`(\d*)d(\d+)([+-]\d+)?`)
	matches := re.FindStringSubmatch(prompt)
	if len(matches) == 0 {
		return nil, "", fmt.Errorf("invalid roll prompt")
	}

	numDice, _ := strconv.Atoi(matches[1])
	if numDice != 2 || matches[2] != "20" {
		return nil, "", fmt.Errorf("fortune only works with 2d20")
	}
	modifier := 0
	if matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	d := dice.NewDice(20)
	details, highest, err := roller.RollWithFortune(d, modifier)
	if err != nil {
		return nil, "", err
	}

	// ✅ FIX: Return details directly (it already includes modifier)
	return []int{highest}, details, nil
}

func parseAndRollWithMisfortune(prompt string) ([]int, string, error) {
	re := regexp.MustCompile(`(\d*)d(\d+)([+-]\d+)?`)
	matches := re.FindStringSubmatch(prompt)
	if len(matches) == 0 {
		return nil, "", fmt.Errorf("invalid roll prompt")
	}

	numDice, _ := strconv.Atoi(matches[1])
	if numDice != 2 || matches[2] != "20" {
		return nil, "", fmt.Errorf("misfortune only works with 2d20")
	}
	modifier := 0
	if matches[3] != "" {
		modifier, _ = strconv.Atoi(matches[3])
	}

	d := dice.NewDice(20)
	details, lowest, err := roller.RollWithMisfortune(d, modifier)
	if err != nil {
		return nil, "", err
	}

	// ✅ FIX: Return details directly (it already includes modifier)
	return []int{lowest}, details, nil
}
