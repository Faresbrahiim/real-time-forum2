export const chatState = {
  currentChatUserId: null,
  currentPage: 1,
  hasMore: false,
  isLoading: false,
  totalMessages: 0,
};
const displayedMessageIds = new Set();
function throttle(func, delay) {
  let timeoutId;
  let lastExecTime = 0;
  return function (...args) {
    const currentTime = Date.now();

    if (currentTime - lastExecTime > delay) {
      func.apply(this, args);
      lastExecTime = currentTime;
    } else {
      clearTimeout(timeoutId);
      timeoutId = setTimeout(() => {
        func.apply(this, args);
        lastExecTime = Date.now();
      }, delay - (currentTime - lastExecTime));
    }
  };
}

export async function startChatWith(userId, username) {
  const chatMessages = document.getElementById("chatMessages");
  const input = document.getElementById("chatInput");
  input.value = "";
  chatMessages.textContent = "";
  displayedMessageIds.clear(); // Reset when opening new chat

  // reset flags becuase we opened a  new convertation
  chatState.currentChatUserId = userId;
  chatState.currentPage = 1;
  chatState.hasMore = false;
  chatState.isLoading = false;
  chatState.totalMessages = 0;

  try {
    // last 10 msgs 
    const response = await fetch(`/api/latest-messages?user_id=${userId}`);
    if (response.ok) {
      const data = await response.json();
      chatState.hasMore = data.has_more;
      chatState.totalMessages = data.total_messages || (data.messages ? data.messages.length : 0);

      displayMessages(data.messages, username, false);
      chatMessages.scrollTop = chatMessages.scrollHeight;
    }
  } catch (err) {
    console.error("Error loading initial messages:", err);
  }

  const chatSection = document.getElementById("chatBox");
  const closeBtn = document.getElementById("closeChat");
  chatSection.style.display = 'flex';
  document.getElementById("chatUsername").textContent = username;
  setupScrollListener(userId, username);

  const newCloseBtn = closeBtn.cloneNode(true);
  closeBtn.parentNode.replaceChild(newCloseBtn, closeBtn);
  newCloseBtn.addEventListener("click", () => {
    chatSection.style.display = 'none';
    const chatMessages = document.getElementById("chatMessages");
    if (chatMessages.scrollHandler) {
      chatMessages.removeEventListener("scroll", chatMessages.scrollHandler);
    }
    chatState.currentChatUserId = null;
  });
}

function setupScrollListener(userId, username) {
  const chatMessages = document.getElementById("chatMessages");

  if (chatMessages.scrollHandler) {
    chatMessages.removeEventListener("scroll", chatMessages.scrollHandler);
  }

  const throttledScrollHandler = throttle(async function () {
    if (chatMessages.scrollTop === 0 && chatState.hasMore && !chatState.isLoading) {
      await loadMoreMessages(userId, username);
    }
  }, 500);

  chatMessages.scrollHandler = throttledScrollHandler;
  chatMessages.addEventListener("scroll", chatMessages.scrollHandler);
}

async function loadMoreMessages(userId, username) {
  if (chatState.isLoading) return;

  chatState.isLoading = true;
  const chatMessages = document.getElementById("chatMessages");
  const loadingDiv = document.createElement("div");
  loadingDiv.className = "loading-indicator";
  loadingDiv.textContent = "Loading more messages...";
  loadingDiv.style.textAlign = "center";
  loadingDiv.style.padding = "10px";
  loadingDiv.style.color = "#666";
  chatMessages.insertBefore(loadingDiv, chatMessages.firstChild);

  try {
    const nextPage = chatState.currentPage + 1;
    const response = await fetch(`/api/messages?user_id=${userId}&page=${nextPage}&limit=10`);

    if (response.ok) {
      const data = await response.json();
      chatMessages.removeChild(loadingDiv);

      if (data.messages && data.messages.length > 0) {
        const oldScrollHeight = chatMessages.scrollHeight;
        displayMessages(data.messages, username, true);
        chatState.currentPage = nextPage;
        chatState.hasMore = data.has_more;

        const newScrollHeight = chatMessages.scrollHeight;
        chatMessages.scrollTop = newScrollHeight - oldScrollHeight;
      } else {
        chatState.hasMore = false;
      }
    }
  } catch (err) {
    console.error("Error loading more messages:", err);
    if (chatMessages.contains(loadingDiv)) {
      chatMessages.removeChild(loadingDiv);
    }
  } finally {
    chatState.isLoading = false;
  }
}

function displayMessages(messages, username, prepend = false) {
  const chatMessages = document.getElementById("chatMessages");
  const currentUserId = chatState.currentChatUserId;

  if (!Array.isArray(messages)) return;

  messages.forEach(msg => {
    if (!msg.id || displayedMessageIds.has(msg.id)) return; // Skip duplicates
    displayedMessageIds.add(msg.id);

    const msgDiv = document.createElement("div");
    msgDiv.className = `message ${msg.from === currentUserId ? "received" : "sent"}`;
    msgDiv.textContent = `${msg.from === currentUserId ? username + ": " : "You: "}${msg.content}`;

    const time = document.createElement("div");
    time.className = "sent-time";
    time.textContent = msg.sent_at;

    msgDiv.appendChild(time);

    if (prepend) {
      chatMessages.insertBefore(msgDiv, chatMessages.firstChild);
    } else {
      chatMessages.appendChild(msgDiv);
    }
  });
}
