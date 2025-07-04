import { ws } from "./script.js";

export const chatState = {
  currentChatUserId: null,
  currentPage: 1,
  hasMore: false,
  isLoading: false,
  totalMessages: 0,
};

// Throttle function to prevent excessive scroll event firing
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
  chatMessages.innerHTML = "";
  
  // Reset chat state
  chatState.currentChatUserId = userId;
  chatState.currentPage = 1;
  chatState.hasMore = false;
  chatState.isLoading = false;
  chatState.totalMessages = 0;

  try {
    // Load initial messages (latest 10)
    const response = await fetch(`/api/latest-messages?user_id=${userId}`);
    if (response.ok) {
      const data = await response.json();
      chatState.hasMore = data.has_more;
      chatState.totalMessages = data.total_messages || data.messages.length;
      
      displayMessages(data.messages, username, false); // false = don't prepend
      
      // Scroll to bottom for initial load
      chatMessages.scrollTop = chatMessages.scrollHeight;
    }
  } catch (err) {
    console.error("Error loading initial messages:", err);
  }

  const chatSection = document.getElementById("chatBox");
  const closeBtn = document.getElementById("closeChat");
  chatSection.style.display = 'flex';
  document.getElementById("chatUsername").textContent = username;

  // Setup scroll listener for pagination
  setupScrollListener(userId, username);
  
  handleTypingToServer(userId);

  const newCloseBtn = closeBtn.cloneNode(true);
  closeBtn.parentNode.replaceChild(newCloseBtn, closeBtn);
  newCloseBtn.addEventListener("click", () => {
    chatSection.style.display = 'none';
    // Remove scroll listener when closing chat
    const chatMessages = document.getElementById("chatMessages");
    if (chatMessages.scrollHandler) {
      chatMessages.removeEventListener("scroll", chatMessages.scrollHandler);
    }
  });
}

function setupScrollListener(userId, username) {
  const chatMessages = document.getElementById("chatMessages");
  
  // Remove existing listener if any
  if (chatMessages.scrollHandler) {
    chatMessages.removeEventListener("scroll", chatMessages.scrollHandler);
  }
  
  // Create throttled scroll handler (300ms delay)
  const throttledScrollHandler = throttle(async function() {
    // Check if user scrolled to top and there are more messages
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
  
  // Show loading indicator
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
      
      // Remove loading indicator
      chatMessages.removeChild(loadingDiv);
      
      if (data.messages && data.messages.length > 0) {
        // Remember scroll position before adding messages
        const oldScrollHeight = chatMessages.scrollHeight;
        
        // Add older messages to the top (load 10 more messages)
        displayMessages(data.messages, username, true); // true = prepend
        
        // Update state
        chatState.currentPage = nextPage;
        chatState.hasMore = data.has_more;
        
        // Maintain scroll position (keep user where they were)
        const newScrollHeight = chatMessages.scrollHeight;
        chatMessages.scrollTop = newScrollHeight - oldScrollHeight;
      } else {
        // No more messages
        chatState.hasMore = false;
      }
    }
  } catch (err) {
    console.error("Error loading more messages:", err);
    // Remove loading indicator on error
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
  
  messages.forEach(msg => {
    const msgDiv = document.createElement("div");
    msgDiv.className = `message ${msg.from === currentUserId ? "received" : "sent"}`;
    msgDiv.textContent = `${msg.from === currentUserId ? username + ": " : "You: "}${msg.content}`;

    const time = document.createElement("div");
    time.className = "sent-time";
    time.textContent = msg.sent_at;

    msgDiv.appendChild(time);
    
    if (prepend) {
      // Add to top for older messages
      chatMessages.insertBefore(msgDiv, chatMessages.firstChild);
    } else {
      // Add to bottom for new messages
      chatMessages.appendChild(msgDiv);
    }
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

// Helper function to add new real-time messages
export function addNewMessage(msg, username) {
  const chatMessages = document.getElementById("chatMessages");
  const msgDiv = document.createElement("div");
  msgDiv.className = "message received";
  msgDiv.textContent = msg.content;

  const time = document.createElement("div");
  time.className = "sent-time";
  time.textContent = msg.sent_at || "";

  msgDiv.appendChild(time);
  chatMessages.appendChild(msgDiv);
  chatMessages.scrollTop = chatMessages.scrollHeight;
}