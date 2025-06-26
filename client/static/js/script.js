// let ws;

// function connectWebSocket() {
//     ws = new WebSocket("ws://localhost:8080/ws");

//     ws.onopen = () => {
//         document.getElementById("chatLog").textContent += "Connected to server\n";
//     };

//     ws.onmessage = (event) => {
//         try {
//             const msg = JSON.parse(event.data);

//             if (msg.type === "user_status") {
//                 displayUserStatus(msg.data);  // Custom function to show users
//             } else {
//                 // Other message types (future support)
//                 document.getElementById("chatLog").textContent += "Server: " + event.data + "\n";
//             }
//         } catch (e) {
//             // Fallback for plain text messages
//             document.getElementById("chatLog").textContent += "Server: " + event.data + "\n";
//         }
//     };

//     ws.onclose = () => {
//         document.getElementById("chatLog").textContent += "Disconnected from server\n";

//         // Optional: Try to reconnect after 3 seconds
//         setTimeout(() => {
//             document.getElementById("chatLog").textContent += "Reconnecting...\n";
//             connectWebSocket();
//         }, 3000);
//     };
// }

// window.sendMessage = function () {
//     const input = document.getElementById("msgInput");
//     if (input.value && ws && ws.readyState === WebSocket.OPEN) {
//         ws.send(input.value);
//         document.getElementById("chatLog").textContent += "You: " + input.value + "\n";
//         input.value = "";
//     }
// };

// // ➕ Function to update the DOM with user status
// function displayUserStatus(users) {
//     const container = document.getElementById("userList");
//     if (!container) return;

//     container.innerHTML = ""; // Clear previous

//     users.forEach(user => {
//         const userDiv = document.createElement("div");
//         userDiv.className = user.online ? "user online" : "user offline";
//         userDiv.textContent = `${user.username} (${user.online ? "🟢" : "⚪"})`;
//         container.appendChild(userDiv);
//     });
// }

// // 🔌 Start WebSocket when the page loads
// window.addEventListener("load", connectWebSocket);

let ws;
function connectWebSocket() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        console.log("WebSocket already connected");
        return;
    }
    ws = new WebSocket("ws://localhost:8080/ws");
    ws.onopen = () => {
        document.getElementById("chatLog").textContent += "✅ Connected to server\n";
    };
    ws.onmessage = (event) => {
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
        document.getElementById("chatLog").textContent += "❌ Disconnected from server\n";

        // Reconnect only if the page is still visible (optional)
        setTimeout(() => {
            document.getElementById("chatLog").textContent += "🔄 Reconnecting...\n";
            connectWebSocket();
        }, 3000);
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
