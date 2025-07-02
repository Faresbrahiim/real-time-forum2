import { ws } from "./script.js";
export const chatState = {
  currentChatUserId: null,
};

export async function startChatWith(userId, username) {
  const chatMessages = document.getElementById("chatMessages");
  chatMessages.innerHTML = "";
  chatState.currentChatUserId = userId;

  try {
    const response = await fetch(`/api/messages?user_id=${userId}`);
    if (response.ok) {
      const messages = await response.json();

      messages.forEach(msg => {
        const msgDiv = document.createElement("div");
        msgDiv.className = `message ${msg.from === userId ? "received" : "sent"}`;
        msgDiv.textContent = `${msg.from === userId ? username +" :" : "You: "}${msg.content}`;

        const time = document.createElement("div");
        time.classList = "sent-time";
        time.textContent = msg.sent_at.slice(11, 16);

        msgDiv.appendChild(time); 
        chatMessages.appendChild(msgDiv);
      });

      chatMessages.scrollTop = chatMessages.scrollHeight;
    }
  } catch (err) {
    
  }

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

