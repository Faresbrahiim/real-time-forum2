export let ws;
function connectWebSocket() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        console.log("WebSocket already connected");
        return;
    }
    ws = new WebSocket("ws://localhost:8080/ws");
    ws.onopen = () => {
        console.log("connect")
    };
    ws.onmessage = (event) => {
        console.log("server msg")
        try {
            const msg = JSON.parse(event.data);

            if (msg.type === "user_status") {
                displayUserStatus(msg.data);
            } else {
                document.getElementById("chatLog").textContent += "Server: " + event.data + "\n";
            }
        } catch (e) {
            document.getElementById("chatLog").textContent += "Server: " + event.data + "\n";
        }
    };
    ws.onclose = () => {
        console.log("disconnect")
    };
}
window.sendMessage = function () {
    const input = document.getElementById("msgInput");
    if (input.value && ws && ws.readyState === WebSocket.OPEN) {
        ws.send(input.value);
        document.getElementById("chatLog").textContent += "You: " + input.value + "\n";
        input.value = "";
    }
};
// ➕ DOM update with online/offline users
function displayUserStatus(users) {
    const container = document.getElementById("userList");
    if (!container) return;

    container.innerHTML = ""; // Clear old

    users.forEach(user => {
        const userDiv = document.createElement("div");
        userDiv.className = `user ${user.online ? "online" : "offline"}`;
        userDiv.textContent = `${user.username} (${user.online ? "🟢" : "⚪"})`;
        container.appendChild(userDiv);
    });
}
// 🔌 Try to connect on page load (or after login)
window.addEventListener("load", () => {
    // Only connect if user is already logged in (optional check via cookie)
    connectWebSocket();
});

// 👉 Call this function *manually* after login via fetch (if using AJAX)
window.afterLogin = function () {
    connectWebSocket();
};
