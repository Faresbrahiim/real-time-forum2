import { ws } from "./script.js";
export const chatState = {
  currentChatUserId: null,
};

export function startChatWith(userId, username) {
  chatState.currentChatUserId = userId;
    
  console.log("Start chat with:", username, "ID:", chatState.currentChatUserId);

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

export function handleTypingIndicator(id, status) {
    const typingIndicator = document.getElementById("typingIndicator");
    if (!typingIndicator) return;
    if (status === "start") {
        typingIndicator.textContent = "Typing...";
        typingIndicator.style.visibility = "visible";
    } else if (status === "stop") {
        typingIndicator.style.visibility = "hidden";
    }
}



