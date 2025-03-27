// Declare WebSocket variable in the global scope
let socket;
let reconnectInterval = 5000; // Reconnect every 5 seconds if disconnected
let reconnectAttempts = 0; // Track the number of reconnection attempts
const maxReconnectAttempts = 5; // Maximum number of reconnection attempts

function initializeWebSocket() {
    socket = new WebSocket("ws://192.168.0.185:8080/ws");

    socket.onopen = () => {
        console.log("WebSocket connection established");
        reconnectAttempts = 0; // Reset reconnection attempts on successful connection
        fetchHistory(); // Fetch roll history once connected
    };

    socket.onmessage = (event) => {
        try {
            const message = JSON.parse(event.data);
            console.log("Message from server:", message);

            if (message.type === "history") {
                console.log("History response received:", message.history);
                const history = document.getElementById("history");
                if (typeof message.history === "string" && message.history.trim() !== "") {
                    const sanitizedHistory = message.history
                        .replace(/\\n/g, "\n")
                        .split("\n")
                        .map((entry) => `<div>${entry}</div>`)
                        .join("");
                    history.innerHTML = sanitizedHistory;
                } else {
                    history.innerHTML = "<div>No roll history available.</div>";
                }
            } else if (message.type === "rollResult") {
                document.getElementById("rollResult").textContent = message.result;
            } else if (message.type === "success") {
                console.log(message.message);
            } else if (message.type === "error") {
                console.error(message.message);
            }
        } catch (error) {
            console.error("Failed to parse server response:", event.data, error);
        }
    };

    socket.onclose = () => {
        console.log("WebSocket connection closed. Attempting to reconnect...");
        if (reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++;
            setTimeout(initializeWebSocket, 5000); // Retry after 5 seconds
        } else {
            console.error("Max reconnection attempts reached. Unable to reconnect.");
        }
    };

    socket.onerror = (error) => {
        console.error("WebSocket error:", error);
    };
}

document.addEventListener("DOMContentLoaded", function () {
    const promptInput = document.getElementById("promptInput");
    let fortuneActive = false;
    let misfortuneActive = false;

    // Initialize WebSocket connection
    initializeWebSocket();

    // Intercept the Enter key globally and trigger the Roll button only when focused on the input field
    document.addEventListener("keydown", function (event) {
        const activeElement = document.activeElement; // Get the currently focused element
        if (event.key === "Enter" && activeElement.id === "promptInput") {
            event.preventDefault(); // Prevent the default behavior of Enter
            document.getElementById("rollButton").click(); // Simulate a click on the Roll button
        }
    });

    // Function to handle button clicks and update the promptInput
    function handleButtonClick(event) {
        const value = event.target.getAttribute("data-value");
        if (!value) return;

        if (value === "clear") {
            // Clear the prompt input
            promptInput.value = "";
        } else if (value === "+") {
            // Add "+" to the prompt input
            promptInput.value += "+";
        } else if (value === "-") {
            // Add "-" to the prompt input
            promptInput.value += "-";
        } else if (value.startsWith("d")) {
            // Handle dice rolls (e.g., d20, d6)
            const currentValue = promptInput.value;
            const regex = new RegExp(`(\\d*)${value}`);
            const match = currentValue.match(regex);

            if (match) {
                // If the die already exists, increment the count
                const count = match[1] ? parseInt(match[1]) + 1 : 2;
                promptInput.value = currentValue.replace(regex, `${count}${value}`);
            } else {
                // If the die doesn't exist, add it to the input
                promptInput.value += value;
            }
        } else {
            // Add the button value (e.g., numbers) to the prompt input
            promptInput.value += value;
        }
    }

    // Attach event listeners to all buttons with data-value attributes
    document.querySelectorAll(".dice-buttons button, .button-grid button, .operator-buttons button").forEach((button) => {
        button.addEventListener("click", handleButtonClick);
    });

    // Handle the Roll button
    document.getElementById("rollButton").addEventListener("click", function () {
        const prompt = promptInput.value;
        if (!prompt) {
            alert("Please enter a roll prompt!");
            return;
        }

        // Determine the roll type based on Fortune/Misfortune
        let rollType = "normal";
        if (fortuneActive) {
            rollType = "fortune";
        } else if (misfortuneActive) {
            rollType = "misfortune";
        }

        console.log(`Roll type: ${rollType}`); // Debug log
        console.log(`Fortune active: ${fortuneActive}, Misfortune active: ${misfortuneActive}`); // Debug log

        if (socket.readyState === WebSocket.OPEN) {
            // Send roll request via WebSocket
            socket.send(
                JSON.stringify({
                    type: "roll",
                    prompt: prompt,
                    rollType: rollType,
                })
            );
        } else {
            console.error("WebSocket is not open. Cannot send roll request.");
        }
    });

    // Handle the Clear History button
    document.getElementById("clearHistoryButton").addEventListener("click", function () {
        if (socket.readyState === WebSocket.OPEN) {
            // Send clear history request via WebSocket
            socket.send(
                JSON.stringify({
                    type: "clear-history",
                })
            );
        } else {
            console.error("WebSocket is not open. Cannot clear history.");
        }
    });

    // Handle Fortune and Misfortune checkboxes
    document.getElementById("fortuneCheckbox").addEventListener("change", function () {
        fortuneActive = this.checked;
        console.log(`Fortune checkbox changed: ${fortuneActive}`); // Debug log
        if (fortuneActive) {
            misfortuneActive = false;
            document.getElementById("misfortuneCheckbox").checked = false;
        }
    });

    document.getElementById("misfortuneCheckbox").addEventListener("change", function () {
        misfortuneActive = this.checked;
        console.log(`Misfortune checkbox changed: ${misfortuneActive}`); // Debug log
        if (misfortuneActive) {
            fortuneActive = false;
            document.getElementById("fortuneCheckbox").checked = false;
        }
    });
});

// Fetch roll history
function fetchHistory() {
    console.log("Fetching roll history...");
    if (socket.readyState === WebSocket.OPEN) {
        console.log("WebSocket is open. Sending history request...");
        socket.send(
            JSON.stringify({
                type: "history",
            })
        );
    } else {
        console.error("WebSocket is not open. Cannot fetch history.");
    }
}