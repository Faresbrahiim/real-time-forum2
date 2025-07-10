import { chatState } from "./chat.js";

export function handleTypingIndicator(msg) {
    const typingDiv = document.getElementById("typingIndicator");
    if (msg.from === chatState.currentChatUserId) {
        typingDiv.style.display = "flex";

        clearTimeout(typingDiv.timeout);
        typingDiv.timeout = setTimeout(() => {
            typingDiv.style.display = "none";
        }, 1500);  // Typing indicator disappears after 1.5 seconds of inactivity
    }
}
