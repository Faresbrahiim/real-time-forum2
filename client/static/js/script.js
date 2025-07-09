export let ws;
import { chatState, startChatWith } from './chat.js';
const unreadNotifs = new Set();
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
                case "message":
                    handleIncomingMessage(msg);
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
    const input = document.getElementById("chatInput");
    const messageText = input.value.trim();
    if (!messageText || !chatState.currentChatUserId || ws.readyState !== WebSocket.OPEN) return;

    ws.send(JSON.stringify({
        type: "message",
        to: chatState.currentChatUserId,
        content: messageText,
    }));

    input.value = "";
};

function displayUserStatus(users) {
    const container = document.getElementById("userList");
    if (!container) return;
    container.innerHTML = "";

    users.forEach(user => {
        const userDiv = document.createElement("div");
        userDiv.className = `user ${user.online ? "online" : "offline"}`;

        const button = document.createElement("button");
        button.className = "username-button";
        button.innerHTML = `${user.online ? "🟢" : "⚪"} ${user.username}`;

        const notifSpan = document.createElement("span");
        notifSpan.id = `notif-${user.id}`;
        notifSpan.textContent = "🔔";

        notifSpan.style.display = unreadNotifs.has(user.id) ? "inline" : "none";

        button.appendChild(notifSpan);
        button.onclick = () => {
            startChatWith(user.id, user.username);
            clearNotif(user.id);
        };

        userDiv.appendChild(button);
        container.appendChild(userDiv);
    });
}
function handleIncomingMessage(msg) {
    let username = "";
     username = document.getElementById("chatUsername").textContent ;
     console.log(username , "username is")
    const chatMessages = document.getElementById("chatMessages");

    if (msg.from === chatState.currentChatUserId) {
        const msgDiv = document.createElement("div");
        msgDiv.className = "message received";
        msgDiv.textContent = username +  ": "+ msg.content;

        const time = document.createElement("div");
        time.className = "sent-time";
        time.textContent = msg.sent_at || "";

        msgDiv.appendChild(time);
        chatMessages.appendChild(msgDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
    else if (msg.receiverId === chatState.currentChatUserId) {
        const msgDiv = document.createElement("div");
        msgDiv.className = "message sent";
        msgDiv.textContent = `You: ${msg.content}`;

        const time = document.createElement("div");
        time.className = "sent-time";
        time.textContent = msg.sent_at || "";

        msgDiv.appendChild(time);
        chatMessages.appendChild(msgDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    }
    else {
        showNotif(msg.from);
        console.log("New message from another user:", msg.from);
    }
}

document.getElementById("sendChat").addEventListener("click", () => {
    window.sendMessage();
});

document.getElementById("chatInput").addEventListener("keydown", (e) => {
    if (e.key === "Enter") {
        window.sendMessage();
    }
});

window.afterLogin = function () {
    connectWebSocket();
};

window.addEventListener('load', async () => {
    const response = await fetch('/api/checksession');
    const result = await response.json();
    if (result.loggedIn) {
        window.currentUsername = result.username; 
        connectWebSocket();
    }
});

function clearNotif(userId) {
    unreadNotifs.delete(userId);
    const notifSpan = document.getElementById(`notif-${userId}`);
    if (notifSpan) {
        notifSpan.style.display = "none";
    }
}

function showNotif(userId) {
    unreadNotifs.add(userId);
    const notifSpan = document.getElementById(`notif-${userId}`);
    if (notifSpan) {
        notifSpan.style.display = "inline";
    }
}
