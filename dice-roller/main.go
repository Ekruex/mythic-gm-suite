package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/Ekruex/mythic-gm-suite/dice-roller/dice"
	"github.com/Ekruex/mythic-gm-suite/dice-roller/roller"
	"github.com/gorilla/websocket"
)

// Struct to represent a WebSocket client with a mutex for synchronized writes
type client struct {
	conn  *websocket.Conn
	mutex sync.Mutex
}

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for now (adjust for production)
	},
}

// Global variables for WebSocket clients and broadcasting
var clients = make(map[*websocket.Conn]*client) // Map to track active WebSocket clients
var clientsMutex sync.Mutex                     // Mutex to synchronize access to the clients map
var broadcast = make(chan string, 100)          // Buffered channel for broadcasting messages

func main() {
	// Start the broadcasting goroutine
	go handleBroadcasts()

	// Pass the broadcast channel to the roller package
	roller.SetBroadcastChannel(broadcast)

	// WebSocket handler
	http.HandleFunc("/ws", handleWebSocket)

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Goroutine to handle broadcasting messages to all connected clients
func handleBroadcasts() {
	for {
		message := <-broadcast
		fmt.Printf("Broadcasting message: %s\n", message) // Debug log
		clientsMutex.Lock()
		for _, c := range clients {
			c.mutex.Lock()
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
			c.mutex.Unlock()
			if err != nil {
				fmt.Println("Error broadcasting message:", err)
				c.conn.Close()
				delete(clients, c.conn)
			}
		}
		clientsMutex.Unlock()
	}
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection:", err)
		return
	}
	defer func() {
		fmt.Println("WebSocket connection closed")
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
		conn.Close()
	}()

	fmt.Println("WebSocket connection established")

	clientsMutex.Lock()
	clients[conn] = &client{conn: conn}
	clientsMutex.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Error reading message: %v\n", err)
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			break
		}
		fmt.Printf("Received message: %s\n", msg)

		response := processWebSocketMessage(string(msg))

		// Synchronize writes to the WebSocket connection
		clientsMutex.Lock()
		c := clients[conn]
		clientsMutex.Unlock()
		c.mutex.Lock()
		err = c.conn.WriteMessage(websocket.TextMessage, []byte(response))
		c.mutex.Unlock()
		if err != nil {
			fmt.Printf("Error writing message: %v\n", err)
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			break
		}
	}
}

func processWebSocketMessage(msg string) string {
	var request map[string]interface{}
	err := json.Unmarshal([]byte(msg), &request)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return `{"type": "error", "message": "Invalid JSON format"}`
	}

	fmt.Printf("Received message: %s\n", msg)

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
			rollType = "normal"
		}
		fmt.Printf("Roll type: %s\n", rollType)

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
		default:
			_, result, err := parseAndRoll(prompt)
			if err != nil {
				return fmt.Sprintf(`{"type": "error", "message": "%s"}`, err.Error())
			}
			return fmt.Sprintf(`{"type": "rollResult", "result": "%s"}`, result)
		}
	case "history":
		history := roller.GetRollHistory()
		fmt.Printf("Retrieved roll history: %v\n", history)

		escapedHistory, err := json.Marshal(strings.Join(history, "\n"))
		if err != nil {
			fmt.Printf("Error marshaling history: %v\n", err)
			return `{"type": "error", "message": "Failed to retrieve roll history"}`
		}

		response := fmt.Sprintf(`{"type": "history", "history": %s}`, string(escapedHistory))
		fmt.Printf("History response: %s\n", response)
		return response
	case "clear-history":
		roller.ClearRollHistory()
		broadcast <- `{"type": "history", "history": ""}`
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
