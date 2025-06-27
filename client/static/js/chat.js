import { ws } from "./script.js";

export function startChatWith(userId, username) {
    console.log("Start chat with:", username, "ID:", userId);

    const chatSection = document.getElementById("chatBox");
    const closeBtn = document.getElementById("closeChat");
    chatSection.style.display = 'flex';
    document.getElementById("chatUsername").textContent = username;

    handleTypingToServer(userId);

    const newCloseBtn = closeBtn.cloneNode(true);
    closeBtn.parentNode.replaceChild(newCloseBtn, closeBtn);
    newCloseBtn.addEventListener("click", () => {
        chatSection.style.display = 'none';
    });
}

function handleTypingToServer(targetUserId) {
    let typingTimeout;
    const input = document.getElementById("chatInput");

    input.addEventListener("keydown", () => {
        console.log("hahuwa dkhul")
        ws.send(JSON.stringify({
            type: "typing",
            to: targetUserId,
            status: "start"
        }));

        clearTimeout(typingTimeout);
        typingTimeout = setTimeout(() => {
            ws.send(JSON.stringify({
                type: "typing",
                to: targetUserId,
                status: "stop"
            }));
        }, 2000);
    });
}

export function handleTypingIndicator(fromUserId, status) {
    const typingIndicator = document.getElementById("typingIndicator");
    if (!typingIndicator) return;

    if (status === "start") {
        typingIndicator.textContent = "Typing...";
        typingIndicator.style.visibility = "visible";
    } else if (status === "stop") {
        typingIndicator.style.visibility = "hidden";
    }
}
