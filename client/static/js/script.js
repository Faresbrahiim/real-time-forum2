export let ws;
import { startChatWith, handleTypingIndicator } from './chat.js';

function connectWebSocket() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        console.log("WebSocket already connected");
        return;
    }

    ws = new WebSocket("ws://localhost:8080/ws");

    ws.onopen = () => {
        console.log("WebSocket connected");
    };

    ws.onmessage = (event) => {
        try {
            const msg = JSON.parse(event.data);
            switch (msg.type) {
                case "user_status":
                    displayUserStatus(msg.data);
                    break;
                case "typing":
                    handleTypingIndicator(msg.from, msg.status);
                    break;
                default:
                    console.log("Unhandled message:", msg);
            }
        } catch (e) {
            console.error("Invalid message:", event.data);
        }
    };

    ws.onclose = () => {
        console.log("WebSocket disconnected");
    };
}

window.sendMessage = function () {
    const input = document.getElementById("msgInput");
    if (input.value && ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            type: "message",
            content: input.value
        }));
        document.getElementById("chatLog").textContent += "You: " + input.value + "\n";
        input.value = "";
    }
};

function displayUserStatus(users) {
    const container = document.getElementById("userList");
    if (!container) return;
    container.innerHTML = "";

    users.forEach(user => {
        const userDiv = document.createElement("div");
        userDiv.className = `user ${user.online ? "online" : "offline"}`;

        const status = document.createElement("div");
        status.className = "status-indicator";
        status.textContent = user.online ? "🟢" : "⚪";

        const button = document.createElement("button");
        button.textContent = user.username;
        button.className = "username-button";
        button.onclick = () => {
            startChatWith(user.id, user.username);
        };

        userDiv.appendChild(status);
        userDiv.appendChild(button);
        container.appendChild(userDiv);
    });
}

window.addEventListener("load", () => {
    connectWebSocket();
});

window.afterLogin = function () {
    connectWebSocket();
};
