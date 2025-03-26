document.addEventListener("DOMContentLoaded", function () {
    const promptInput = document.getElementById("promptInput");
    let fortuneActive = false;
    let misfortuneActive = false;

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

        fetch(`/api/roll?prompt=${encodeURIComponent(prompt)}&type=${rollType}`)
            .then((response) => {
                if (!response.ok) {
                    throw new Error(`Server error: ${response.status}`);
                }
                return response.text();
            })
            .then((result) => {
                document.getElementById("rollResult").textContent = result;
                // Fetch the updated history after the roll
            fetchHistory();
            })
            
            .catch((error) => {
                console.error("Error fetching roll result:", error);
                alert("An error occurred while fetching the roll result. Please try again.");
            });
    });

    // Handle the Clear History button
    document.getElementById("clearHistoryButton").addEventListener("click", function () {
        // Clear the roll history
        fetch("/api/clear-history", { method: "POST" })
            .then(() => {
                document.getElementById("history").textContent = "History cleared.";
            })
            .catch((error) => {
                console.error("Error clearing history:", error);
            });
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

    // Fetch roll history on page load
    function fetchHistory() {
        fetch("/api/history")
            .then((response) => response.text())
            .then((data) => {
                const history = document.getElementById("history");
                history.innerHTML = data.split("\n").map((entry) => `<div>${entry}</div>`).join("");
            });
    }

    fetchHistory();
});