package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	g "forum/server/global"
	session "forum/server/session"

	"github.com/gorilla/websocket"
)

type TypingMessage struct {
	Type   string `json:"type"`
	To     string `json:"to"`
	Status string `json:"status"`
}

type MessagePayload struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Content string `json:"content"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, err := session.GetSessionUserID(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"status": %d, "message": "You must be logged in"}`, http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()
	g.ActiveConnectionsMutex.Lock()
	g.ActiveConnections[userID] = &g.SafeConn{Conn: conn}
	g.ActiveConnectionsMutex.Unlock()

	BroadcastUserStatus()

	defer func() {
		g.ActiveConnectionsMutex.Lock()
		delete(g.ActiveConnections, userID)
		g.ActiveConnectionsMutex.Unlock()

		BroadcastUserStatus()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		var base map[string]interface{}
		if err := json.Unmarshal(msg, &base); err != nil {
			log.Println("Invalid JSON from client")
			continue
		}

		switch base["type"] {
		case "message":
			{
				handleIncomingMessage(userID, msg)
			}
		default:
			fmt.Printf("Received unhandled: %s\n", msg)
		}
	}
}

func GetOnlineUsers(_ string) (map[string]bool, map[string]string) {
	onlineUsers := make(map[string]bool)
	allUsers := make(map[string]string)

	rows, err := g.DB.Query("SELECT id, username FROM users")
	if err != nil {
		log.Println("Error selecting users:", err)
		return nil, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Println("Error scanning user:", err)
			continue
		}
		onlineUsers[id] = false
		allUsers[id] = name
	}
	g.ActiveConnectionsMutex.RLock()
	for id := range g.ActiveConnections {
		onlineUsers[id] = true
	}
	g.ActiveConnectionsMutex.RUnlock()

	return onlineUsers, allUsers
}

func BroadcastUserStatus() {
	onlineUsers, allUsers := GetOnlineUsers("")
	if onlineUsers == nil || allUsers == nil {
		return
	}

	var userList []map[string]interface{}
	for id, name := range allUsers {
		userList = append(userList, map[string]interface{}{
			"id":       id,
			"username": name,
			"online":   onlineUsers[id],
		})
	}

	update := map[string]interface{}{
		"type": "user_status",
		"data": userList,
	}

	BroadcastToAllUsers(update)
}

func BroadcastToAllUsers(update interface{}) {
	jsonUpdate, err := json.Marshal(update)
	if err != nil {
		log.Println("Error marshaling update:", err)
		return
	}

	g.ActiveConnectionsMutex.RLock()
	defer g.ActiveConnectionsMutex.RUnlock()

	for userID, conn := range g.ActiveConnections {
		conn.WriteMu.Lock()
		err := conn.Conn.WriteMessage(websocket.TextMessage, jsonUpdate)
		conn.WriteMu.Unlock()

		if err != nil {
			log.Println("Error writing to user", userID, ":", err)
		}
	}
}



func handleIncomingMessage(senderID string, msg []byte) {
	var payload MessagePayload
	if err := json.Unmarshal(msg, &payload); err != nil {
		log.Println("Invalid message payload:", err)
		return
	}

	receiverID := payload.To
	content := payload.Content

	convoID, err := getOrCreateConversation(senderID, receiverID)
	if err != nil {
		log.Println("Error getting/creating conversation:", err)
		return
	}
	_, err = g.DB.Exec(`
		INSERT INTO Messages (id, conversation_id, sender_id, receiver_id, content, seen)
		VALUES (?, ?, ?, ?, ?, ?)`,
		g.GenerateUUID(), convoID, senderID, receiverID, content, false)
	if err != nil {
		log.Println("Error inserting message:", err)
		return
	}
	deliverMessageToUser(senderID, receiverID, content, convoID)
	
}

func getOrCreateConversation(user1, user2 string) (string, error) {
	var convoID string


	query := `
		SELECT id FROM Conversations 
		WHERE (user1_id = ? AND user2_id = ?) 
		   OR (user1_id = ? AND user2_id = ?)
	`
	err := g.DB.QueryRow(query, user1, user2, user2, user1).Scan(&convoID)
	if err == nil {
		return convoID, nil
	}

	convoID = g.GenerateUUID()
	_, err = g.DB.Exec(`
		INSERT INTO Conversations (id, user1_id, user2_id)
		VALUES (?, ?, ?)`,
		convoID, user1, user2)
	if err != nil {
		return "", err
	}

	return convoID, nil
}

func deliverMessageToUser(senderID, receiverID, content, conversationID string) {
	g.ActiveConnectionsMutex.RLock()
	receiverConn, receiverOnline := g.ActiveConnections[receiverID]
	senderConn, senderOnline := g.ActiveConnections[senderID]
	g.ActiveConnectionsMutex.RUnlock()

	messagePayload := map[string]interface{}{
		"type":            "message",
		"from":            senderID,
		"content":         content,
		"receiverId":      receiverID,
		"conversation_id": conversationID,
		"sent_at":         time.Now().Format(time.RFC3339),
	}

	jsonMsg, err := json.Marshal(messagePayload)
	if err != nil {
		log.Println("Error marshaling message to deliver:", err)
		return
	}

	if receiverOnline {
		receiverConn.WriteMu.Lock()
		err := receiverConn.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
		receiverConn.WriteMu.Unlock()
		if err != nil {
			log.Println("Error sending message to receiver:", err)
		}
	}

	// Optionally send back to sender (to confirm/send in UI)
	if senderOnline {
		senderConn.WriteMu.Lock()
		err := senderConn.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
		senderConn.WriteMu.Unlock()
		if err != nil {
			log.Println("Error sending message back to sender:", err)
		}
	}
}

func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := session.GetSessionUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	targetID := r.URL.Query().Get("user_id")
	if targetID == "" {
		http.Error(w, "Missing user_id", http.StatusBadRequest)
		return
	}

	convoID, err := getOrCreateConversation(userID, targetID)
	fmt.Println("the conver id id", convoID)

	if err != nil {
		http.Error(w, "Failed to get conversation", http.StatusInternalServerError)
		return
	}
	rows, err := g.DB.Query(`
	SELECT sender_id, receiver_id, content, sent_at 
	FROM Messages 
	WHERE conversation_id = ? 
	ORDER BY sent_at ASC
	`, convoID)
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type Message struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Content string `json:"content"`
		SentAt  string `json:"sent_at"`
	}

	var messages []Message

	for rows.Next() {

		var m Message
		fmt.Println("sss", m)
		var sentTime time.Time
		if err := rows.Scan(&m.From, &m.To, &m.Content, &sentTime); err != nil {
			log.Println("Error scanning message row:", err)
			continue
		}
		m.SentAt = sentTime.Format(time.RFC3339)
		messages = append(messages, m)
	}
	
	fmt.Println("messages are", messages)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
