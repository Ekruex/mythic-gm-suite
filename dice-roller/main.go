package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
	"github.com/Ekruex/mythic-gm-suite/dice-roller/roller"
	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now (adjust for production)
	},
}

func main() {
	// WebSocket handler
	http.HandleFunc("/ws", handleWebSocket)

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer conn.Close()

	fmt.Println("WebSocket connection established")

	for {
		// Read message from client
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}
		fmt.Printf("Received: %s\n", msg)

		// Process the message
		response := processWebSocketMessage(string(msg))

		// Send response back to client
		err = conn.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil {
			fmt.Println("Error writing message:", err)
			break
		}
	}
}

func processWebSocketMessage(msg string) string {
	// Parse the message as JSON
	var request map[string]interface{}
	err := json.Unmarshal([]byte(msg), &request)
	if err != nil {
		return `{"type": "error", "message": "Invalid JSON format"}`
	}

	// Log the received message
	fmt.Printf("Received message: %s\n", msg)

	// Handle different message types
	switch request["type"] {
	case "roll":
		prompt, ok := request["prompt"].(string)
		if !ok {
			return `{"type": "error", "message": "Invalid roll prompt"}`
		}
		if prompt == "" {
			return `{"type": "error", "message": "Roll prompt cannot be empty"}`
		}

		rollType, ok := request["rollType"].(string)
		if !ok {
			rollType = "normal" // Default to "normal" if rollType is missing
		}
		fmt.Printf("Roll type: %s\n", rollType) // Log the rollType

		// Handle roll types: normal, fortune, misfortune
		switch rollType {
		case "fortune":
			_, result, err := parseAndRollWithFortune(prompt)
			if err != nil {
				return fmt.Sprintf(`{"type": "error", "message": "%s"}`, err.Error())
			}
			return fmt.Sprintf(`{"type": "rollResult", "result": "%s"}`, result)
		case "misfortune":
			_, result, err := parseAndRollWithMisfortune(prompt)
			if err != nil {
				return fmt.Sprintf(`{"type": "error", "message": "%s"}`, err.Error())
			}
			return fmt.Sprintf(`{"type": "rollResult", "result": "%s"}`, result)
		default: // Normal roll
			_, result, err := parseAndRoll(prompt)
			if err != nil {
				return fmt.Sprintf(`{"type": "error", "message": "%s"}`, err.Error())
			}
			return fmt.Sprintf(`{"type": "rollResult", "result": "%s"}`, result)
		}
	case "history":
		history := roller.GetRollHistory()
		fmt.Println("Processing history request")

		// Use json.Marshal to escape special characters, including newlines
		escapedHistory, err := json.Marshal(strings.Join(history, "\n"))
		if err != nil {
			fmt.Printf("Error marshaling history: %v\n", err)
			return `{"type": "error", "message": "Failed to retrieve roll history"}`
		}

		// Properly format the JSON response
		return fmt.Sprintf(`{"type": "history", "history": %s}`, string(escapedHistory))
	case "clear-history":
		roller.ClearRollHistory()
		return `{"type": "success", "message": "Roll history cleared"}`
	default:
		return `{"type": "error", "message": "Unknown message type"}`
	}
}

// functions for dice rolling logic
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

	return []int{lowest}, details, nil
}
