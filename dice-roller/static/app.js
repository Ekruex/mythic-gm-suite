document.addEventListener("DOMContentLoaded", function() {
    const root = document.getElementById("root");
    let fortuneActive = false;
    let misfortuneActive = false;

    function fetchRoll(prompt) {
        let url = `/api/roll?prompt=${encodeURIComponent(prompt)}`;
        if (fortuneActive) {
            url = `/api/fortune?prompt=${encodeURIComponent(prompt)}`;
        } else if (misfortuneActive) {
            url = `/api/misfortune?prompt=${encodeURIComponent(prompt)}`;
        }

        fetch(url)
            .then(response => response.text())
            .then(data => {
                const rollResult = document.getElementById("rollResult");
                rollResult.textContent = data;
                fetchHistory();
            });
    }

    function fetchHistory() {
        fetch("/api/history")
            .then(response => response.text())
            .then(data => {
                const history = document.getElementById("history");
                history.innerHTML = data.split('\n').map(entry => `<div>${entry}</div>`).join('');
            });
    }

    function clearHistory() {
        fetch("/api/clear-history", { method: "POST" })
            .then(() => {
                fetchHistory();
            });
    }

    root.innerHTML = `
        <h1>Mythic GM Suite</h1>
        <div class="button-grid">
            <button data-value="1">1</button>
            <button data-value="2">2</button>
            <button data-value="3">3</button>
            <button data-value="4">4</button>
            <button data-value="5">5</button>
            <button data-value="6">6</button>
            <button data-value="7">7</button>
            <button data-value="8">8</button>
            <button data-value="9">9</button>
            <button data-value="0">0</button>
            <button data-value="d4">d4</button>
            <button data-value="d6">d6</button>
            <button data-value="d8">d8</button>
            <button data-value="d10">d10</button>
            <button data-value="d12">d12</button>
            <button data-value="d20">d20</button>
            <button data-value="d100">d100</button>
            <button data-value="+">+</button>
            <button data-value="-">-</button>
            <button data-value="clear">Clear</button>
            <button id="rollButton">Roll</button>
            <button id="clearHistoryButton">Clear History</button>
        </div>
        <div class="checkbox-grid">
            <label>
                <input type="checkbox" id="fortuneCheckbox">
                Fortune
            </label>
            <label>
                <input type="checkbox" id="misfortuneCheckbox">
                Misfortune
            </label>
        </div>
        <input type="text" id="promptInput" placeholder="Enter roll prompt (e.g., 5d6+12)">
        <div id="rollResult"></div>
        <h2>Roll History</h2>
        <div id="history"></div>
    `;

    const promptInput = document.getElementById("promptInput");

    document.querySelectorAll(".button-grid button").forEach(button => {
        button.addEventListener("click", function() {
            const value = this.getAttribute("data-value");
            if (value === "clear") {
                promptInput.value = "";
            } else if (value === "Roll") {
                fetchRoll(promptInput.value);
            } else {
                // Handle multiple dice of the same type
                const currentValue = promptInput.value;
                if (value.startsWith("d")) {
                    const regex = new RegExp(`(\\d*)${value}`);
                    const match = currentValue.match(regex);
                    if (match) {
                        const count = match[1] ? parseInt(match[1]) + 1 : 2;
                        promptInput.value = currentValue.replace(regex, `${count}${value}`);
                    } else {
                        promptInput.value += value;
                    }
                } else if (value === "+" || value === "-") {
                    // Ensure + or - is added correctly after dice
                    const lastChar = currentValue.slice(-1);
                    if (lastChar === "+" || lastChar === "-") {
                        promptInput.value = currentValue.slice(0, -1) + value;
                    } else {
                        promptInput.value += value;
                    }
                } else {
                    promptInput.value += value;
                }
            }
        });
    });

    document.getElementById("rollButton").addEventListener("click", function() {
        fetchRoll(promptInput.value);
    });

    document.getElementById("clearHistoryButton").addEventListener("click", function() {
        clearHistory();
    });

    document.getElementById("fortuneCheckbox").addEventListener("change", function() {
        fortuneActive = this.checked;
        if (fortuneActive) {
            misfortuneActive = false;
            document.getElementById("misfortuneCheckbox").checked = false;
        }
    });

    document.getElementById("misfortuneCheckbox").addEventListener("change", function() {
        misfortuneActive = this.checked;
        if (misfortuneActive) {
            fortuneActive = false;
            document.getElementById("fortuneCheckbox").checked = false;
        }
    });

    fetchHistory();
});