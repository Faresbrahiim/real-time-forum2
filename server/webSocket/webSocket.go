package ws

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	g "forum/server/global"
	session "forum/server/session"

	"github.com/gorilla/websocket"
)

type MessagePayload struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Content string `json:"content"`
}

type UserStatus struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

type Message struct {
	ID      int64  `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
	SentAt  string `json:"sent_at"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:8080"
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

	safeConn := &g.SafeConn{Conn: conn}

	g.ActiveConnectionsMutex.Lock()
	g.ActiveConnections[userID] = append(g.ActiveConnections[userID], safeConn)
	g.ActiveConnectionsMutex.Unlock()

	BroadcastUserStatus()

	defer func() {
		g.ActiveConnectionsMutex.Lock()
		conns := g.ActiveConnections[userID]
		for i, c := range conns {
			if c == safeConn {
				// Remove this connection
				g.ActiveConnections[userID] = append(conns[:i], conns[i+1:]...)
				break
			}
		}
		// Clean up if no connections left
		if len(g.ActiveConnections[userID]) == 0 {
			delete(g.ActiveConnections, userID)
		}
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
			handleIncomingMessage(userID, msg)
		default:
			fmt.Printf("Received unhandled: %s\n", msg)
		}
	}
}

func BroadcastUserStatus() {
	g.ActiveConnectionsMutex.RLock()
	defer g.ActiveConnectionsMutex.RUnlock()

	for userID, conns := range g.ActiveConnections {
		users := GetOnlineUsers(userID)
		if users == nil {
			continue
		}

		update := map[string]interface{}{
			"type": "user_status",
			"data": users,
		}

		jsonUpdate, err := json.Marshal(update)
		if err != nil {
			log.Println("Error marshaling update:", err)
			continue
		}

		for _, conn := range conns {
			conn.WriteMu.Lock()
			conn.Conn.WriteMessage(websocket.TextMessage, jsonUpdate)
			conn.WriteMu.Unlock()
		}
	}
}

func UpdateUserStatusForUsers(userIDs []string) {
	g.ActiveConnectionsMutex.RLock()
	defer g.ActiveConnectionsMutex.RUnlock()

	for _, targetUserID := range userIDs {
		conns, exists := g.ActiveConnections[targetUserID]
		if !exists {
			continue
		}

		users := GetOnlineUsers(targetUserID)
		if users == nil {
			continue
		}

		update := map[string]interface{}{
			"type": "user_status",
			"data": users,
		}

		jsonUpdate, err := json.Marshal(update)
		if err != nil {
			log.Println("Error marshaling update:", err)
			continue
		}

		for _, conn := range conns {
			conn.WriteMu.Lock()
			conn.Conn.WriteMessage(websocket.TextMessage, jsonUpdate)
			conn.WriteMu.Unlock()
		}
	}
}

func GetOnlineUsers(excludeUserID string) []UserStatus {
	var users []UserStatus

	query := `
		SELECT u.id, u.username, MAX(m.sent_at) AS last_message
		FROM users u
		LEFT JOIN (
			SELECT sender_id AS user_id, sent_at FROM Messages
			UNION ALL
			SELECT receiver_id AS user_id, sent_at FROM Messages
		) m ON u.id = m.user_id
		WHERE u.id != ?
		GROUP BY u.id
		ORDER BY
			CASE WHEN last_message IS NULL THEN 1 ELSE 0 END,
			last_message DESC,
			u.username ASC
	`

	rows, err := g.DB.Query(query, excludeUserID)
	if err != nil {
		log.Println("Error selecting users:", err)
		return nil
	}
	defer rows.Close()

	onlineMap := make(map[string]bool)
	g.ActiveConnectionsMutex.RLock()
	for id := range g.ActiveConnections {
		onlineMap[id] = true
	}
	g.ActiveConnectionsMutex.RUnlock()

	for rows.Next() {
		var id, username string
		var lastMessage sql.NullString

		if err := rows.Scan(&id, &username, &lastMessage); err != nil {
			log.Println("Error scanning user:", err)
			continue
		}

		users = append(users, UserStatus{
			ID:       id,
			Username: username,
			Online:   onlineMap[id],
		})
	}

	return users
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
		INSERT INTO Messages (conversation_id, sender_id, receiver_id, content, seen)
		VALUES (?, ?, ?, ?, ?)`, convoID, senderID, receiverID, content, false)
	if err != nil {
		log.Println("Error inserting message:", err)
		return
	}

	deliverMessageToUser(senderID, receiverID, content, convoID)
	UpdateUserStatusForUsers([]string{senderID, receiverID})
}

func getOrCreateConversation(user1, user2 string) (int64, error) {
	var convoID int64
	err := g.DB.QueryRow(`
		SELECT id FROM Conversations 
		WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)
	`, user1, user2, user2, user1).Scan(&convoID)
	if err == nil {
		return convoID, nil
	}

	res, err := g.DB.Exec(`
		INSERT INTO Conversations (user1_id, user2_id) VALUES (?, ?)
	`, user1, user2)
	if err != nil {
		return 0, err
	}

	convoID, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return convoID, nil
}

func deliverMessageToUser(senderID, receiverID, content string, conversationID int64) {
	g.ActiveConnectionsMutex.RLock()
	receiverConns := g.ActiveConnections[receiverID]
	senderConns := g.ActiveConnections[senderID]
	g.ActiveConnectionsMutex.RUnlock()

	messagePayload := map[string]interface{}{
		"type":            "message",
		"from":            senderID,
		"content":         content,
		"receiverId":      receiverID,
		"conversation_id": conversationID,
		"sent_at":         time.Now().Format("15:04"),
	}

	jsonMsg, err := json.Marshal(messagePayload)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	for _, conn := range senderConns {
		conn.WriteMu.Lock()
		conn.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
		conn.WriteMu.Unlock()
	}

	for _, conn := range receiverConns {
		conn.WriteMu.Lock()
		conn.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
		conn.WriteMu.Unlock()
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

	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}

	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		http.Error(w, "Invalid page number", http.StatusBadRequest)
		return
	}

	limit := 10
	offset := (pageNum - 1) * limit

	convoID, err := getOrCreateConversation(userID, targetID)
	if err != nil {
		http.Error(w, "Failed to get conversation", http.StatusInternalServerError)
		return
	}

	var totalMessages int
	err = g.DB.QueryRow("SELECT COUNT(*) FROM Messages WHERE conversation_id = ?", convoID).Scan(&totalMessages)
	if err != nil {
		http.Error(w, "Failed to count messages", http.StatusInternalServerError)
		return
	}

	rows, err := g.DB.Query(`
		SELECT id, sender_id, receiver_id, content, sent_at
		FROM Messages
		WHERE conversation_id = ?
		ORDER BY id DESC
		LIMIT ? OFFSET ?`, convoID, limit, offset)
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var sentTime time.Time
		if err := rows.Scan(&m.ID, &m.From, &m.To, &m.Content, &sentTime); err != nil {
			log.Println("Error scanning message row:", err)
			continue
		}
		m.SentAt = sentTime.Add(1 * time.Hour).Format("15:04")
		messages = append(messages, m)
	}

	totalPages := (totalMessages + limit - 1) / limit
	hasMore := pageNum < totalPages

	response := map[string]interface{}{
		"messages":       messages,
		"current_page":   pageNum,
		"total_pages":    totalPages,
		"total_messages": totalMessages,
		"has_more":       hasMore,
		"per_page":       limit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetLatestMessagesHandler(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		http.Error(w, "Failed to get conversation", http.StatusInternalServerError)
		return
	}

	rows, err := g.DB.Query(`
		SELECT id, sender_id, receiver_id, content, sent_at
		FROM Messages
		WHERE conversation_id = ?
		ORDER BY id DESC
		LIMIT 10`, convoID)
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		var sentTime time.Time
		if err := rows.Scan(&m.ID, &m.From, &m.To, &m.Content, &sentTime); err != nil {
			log.Println("Error scanning message row:", err)
			continue
		}
		m.SentAt = sentTime.Add(1 * time.Hour).Format("15:04")
		messages = append(messages, m)
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	var hasMore bool
	err = g.DB.QueryRow("SELECT COUNT(*) > 10 FROM Messages WHERE conversation_id = ?", convoID).Scan(&hasMore)
	if err != nil {
		log.Println("Error checking for more messages:", err)
		hasMore = false
	}

	response := map[string]interface{}{
		"messages": messages,
		"has_more": hasMore,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
