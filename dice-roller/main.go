package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

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
		return true // Allow all origins for now
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

	// HTTP handlers
	http.HandleFunc("/history", roller.HandleFetchHistory) // Fetch roll history
	http.HandleFunc("/roll", roller.HandleRoll)            // Roll dice

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	// Start HTTP server on port 8080
	fmt.Println("Server is running on http://0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Goroutine to handle broadcasting messages to all connected clients
func handleBroadcasts() {
	for {
		message := <-broadcast
		clientsMutex.Lock()
		for _, c := range clients {
			c.mutex.Lock()
			err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
			c.mutex.Unlock()
			if err != nil {
				c.conn.Close()
				delete(clients, c.conn)
			}
		}
		clientsMutex.Unlock()
	}
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Log the incoming WebSocket request headers for debugging
	log.Printf("WebSocket request headers: %+v\n", r.Header)

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Log the error and return an HTTP error response
		log.Printf("Failed to upgrade WebSocket: %v\n", err)
		http.Error(w, "Failed to upgrade WebSocket", http.StatusBadRequest)
		return
	}
	log.Println("WebSocket connection successfully upgraded")

	defer func() {
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
		conn.Close()
		log.Println("WebSocket connection closed")
	}()

	// Add the WebSocket connection to the clients map
	clientsMutex.Lock()
	clients[conn] = &client{conn: conn}
	clientsMutex.Unlock()
	log.Println("WebSocket client added to the clients map")

	// Handle incoming WebSocket messages
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v\n", err)
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			break
		}
		log.Printf("Received WebSocket message: %s\n", string(msg))

		// Process the WebSocket message
		response := processWebSocketMessage(string(msg))

		// Send the response back to the client
		clientsMutex.Lock()
		c := clients[conn]
		clientsMutex.Unlock()
		c.mutex.Lock()
		err = c.conn.WriteMessage(websocket.TextMessage, []byte(response))
		c.mutex.Unlock()
		if err != nil {
			log.Printf("WebSocket write error: %v\n", err)
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			break
		}
		log.Printf("Sent WebSocket response: %s\n", response)
	}
}

func processWebSocketMessage(msg string) string {
	var request map[string]interface{}
	err := json.Unmarshal([]byte(msg), &request)
	if err != nil {
		return `{"type": "error", "message": "Invalid JSON format"}`
	}

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

		switch rollType {
		case "fortune":
			_, result, err := roller.ParseAndRollWithFortune(prompt)
			if err != nil {
				return fmt.Sprintf(`{"type": "error", "message": "%s"}`, err.Error())
			}
			return fmt.Sprintf(`{"type": "rollResult", "result": "%s"}`, result)
		case "misfortune":
			_, result, err := roller.ParseAndRollWithMisfortune(prompt)
			if err != nil {
				return fmt.Sprintf(`{"type": "error", "message": "%s"}`, err.Error())
			}
			return fmt.Sprintf(`{"type": "rollResult", "result": "%s"}`, result)
		default:
			_, result, err := roller.ParseAndRoll(prompt)
			if err != nil {
				return fmt.Sprintf(`{"type": "error", "message": "%s"}`, err.Error())
			}
			return fmt.Sprintf(`{"type": "rollResult", "result": "%s"}`, result)
		}
	case "history":
		history := roller.GetRollHistory()
		escapedHistory, err := json.Marshal(strings.Join(history, "\n"))
		if err != nil {
			return `{"type": "error", "message": "Failed to retrieve roll history"}`
		}
		return fmt.Sprintf(`{"type": "history", "history": %s}`, string(escapedHistory))
	case "clear-history":
		roller.ClearRollHistory()
		broadcast <- `{"type": "history", "history": ""}`
		return `{"type": "success", "message": "Roll history cleared"}`
	default:
		return `{"type": "error", "message": "Unknown message type"}`
	}
}
