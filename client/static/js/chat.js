import { ws } from "./script.js";

export function startChatWith(userId, username) {

    console.log("Start chat with:", username, "ID:", userId);

    const chatSection = document.getElementById("chatBox");
    const closeBtn = document.getElementById("closeChat");
    // display flex because it was display none...
    chatSection.style.display = 'flex';
    // the  header of chat have username
    document.getElementById("chatUsername").textContent = username;
    handleTypingToServer(userId);
    // clone node is methose to create a new  node element true means deepclone false means clone only the elemnt not its children
    const newCloseBtn = closeBtn.cloneNode(true);
    // recplace child ,....  bayna men smiytha 
    closeBtn.parentNode.replaceChild(newCloseBtn, closeBtn);
    newCloseBtn.addEventListener("click", () => {
        // chatsection  removed
        chatSection.style.display = 'none';
    });
    // Cloning and replacing the element is a quick way to remove all previously attached event listeners on the button without manually tracking them

}

function handleTypingToServer(targetUserId) {
    let typingTimeout;
    /// chat input .... 
    const input = document.getElementById("chatInput");

    input.addEventListener("keydown", () => {
        // console.log("hahuwa dkhul")
        // send to server a msg ..... chosen freind and typing is type and statut is start 
        ws.send(JSON.stringify({
            type: "typing",
            to: targetUserId,
            status: "start"
        }));

        clearTimeout(typingTimeout);
        // set time out to check  if user is stopped typing  ... .. indirect method each 2 scds
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
    // if status of typing type is start !! 
    if (status === "start") {
        typingIndicator.textContent = "Typing...";
        typingIndicator.style.visibility = "visible";
    } else if (status === "stop") {
        typingIndicator.style.visibility = "hidden";
    }
}
