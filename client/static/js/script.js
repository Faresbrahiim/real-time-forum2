export let ws;

import { chatState, startChatWith, handleTypingIndicator } from './chat.js';

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
        content: messageText
    }));


    const chatMessages = document.getElementById("chatMessages");
    const msgDiv = document.createElement("div");
    msgDiv.className = "message sent";
    msgDiv.textContent = `You: ${messageText}`;
    chatMessages.appendChild(msgDiv);

    
    chatMessages.scrollTop = chatMessages.scrollHeight;
    input.value = "";
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


window.afterLogin = function () {
    connectWebSocket();
};


function handleIncomingMessage(msg) {
    console.log("Incoming:", msg);

    if (msg.from === chatState.currentChatUserId) {
        console.log("hhhhhhhh",chatState.currentChatUserId);
        
        const chatMessages = document.getElementById("chatMessages");
        const msgDiv = document.createElement("div");
        msgDiv.className = "message received";
        msgDiv.textContent = msg.content;
        chatMessages.appendChild(msgDiv);
        chatMessages.scrollTop = chatMessages.scrollHeight;
    } else {
        console.log("New message from another user:", msg.from);
        // Optional: show notification or badge here
    }
}





document.getElementById("sendChat").addEventListener("click", () => {
    window.sendMessage();
});


///
