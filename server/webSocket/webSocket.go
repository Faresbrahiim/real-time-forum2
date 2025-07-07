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

// msg playlod is used to send a msg to the client such as  in content mssgs ... typing user_status etc....
type MessagePayload struct {
	Type    string `json:"type"`
	To      string `json:"to"`
	Content string `json:"content"`
}

// userStatus struct user to store the status of the user such as id username and online offline
type UserStatus struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Online   bool   `json:"online"`
}

// msgs is used to send  msg to  clinet with it content .. sender receiver  ,,, sending time and the id ,,,
type Message struct {
	ID      int64  `json:"id"`
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
	SentAt  string `json:"sent_at"`
}

// upgrade struch used to upgrade the http request to websocket connection  ,, check orgigin only upgrage when the request comes from the last origin to avoid attack ..
// why http://localhost:8080  exactly ? because the backend server serve both the static files such as js html...(front) and the back also
// when we used react or something else the server origin is http://localhost:3030
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:8080"
	},
}

// WebSocket Handler wait for a request from the front when user get logged  .. route  => /ws
// check if the session exist on db if no return json as response .... if not continue ans upgrade the connection
// defer conn.close()  execute when the  function finish ,,,, it clone the connection
// seconde defer used to delte the connection from the  active connection ,, given it the id
// which defer it will execute first ... the secone one why ? because the defer func pusheed into stack which use the lifo mecanisme
// after upgrading to webSocket connection means that ,, the user  or the current connection is online  we marked it by adding the user if and the connection in the map
//
//	why mutex ? to avoid cunccurency and race condition since the map is shared beteween many connections or users that means ,,, we have to lock and unclock after any action
//
// defer ... exxetuce in cases  of error returning or something...,,, like refresh page close connection in general
// the for loop is lesnting to msgs from front
// reads the msg from the connection _ msg type ... msg ... byte slice
// declare a map to holde parsed json
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
			handleIncomingMessage(userID, msg)
		default:
			fmt.Printf("Received unhandled: %s\n", msg)
		}
	}
}

// after a msg is sent ... means the last msg time is change the order of msg changes means the order of users change... means..., update it for the two users
// for the receiver the the sender ..., do the same query again but only for the two users...
// the connection will be for both the sender and the receiver because handleIncomming msg handled them both
func UpdateUserStatusForUsers(userIDs []string) {
	g.ActiveConnectionsMutex.RLock()
	defer g.ActiveConnectionsMutex.RUnlock()

	for _, targetUserID := range userIDs {
		conn, exists := g.ActiveConnections[targetUserID]
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

		conn.WriteMu.Lock()
		err = conn.Conn.WriteMessage(websocket.TextMessage, jsonUpdate)
		conn.WriteMu.Unlock()

		if err != nil {
			log.Println("Error writing to user", targetUserID, ":", err)
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

// brodcast  the status to all active connection ....
// RLock because we will just read no write
// for each userid we get the online users ... except the current one to not show it in ui
// if  none on users is online just continue
// update is a map that accpet string in key and interface in value can be any type....
// type is user_statue so the front  know how to handle it
// the data is users.... online with true offline with false
// json marshal ... make the map => to json format since  josn is convenient choice.
// since we're writing... we need to make the connection safe means we will use mutex
// why we did not call g global because conn is like an instance of the  safeconncetion as we said before
// textMessage --> send a msg as plein text not binary or somthing...
func BroadcastUserStatus() {
	g.ActiveConnectionsMutex.RLock()
	defer g.ActiveConnectionsMutex.RUnlock()

	for userID, conn := range g.ActiveConnections {
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

		conn.WriteMu.Lock()
		err = conn.Conn.WriteMessage(websocket.TextMessage, jsonUpdate)
		conn.WriteMu.Unlock()

		if err != nil {
			log.Println("Error writing to user", userID, ":", err)
		}
	}
}

// in case of type is  message ,,, with the msg and  the sender ...
// the args is slice of byte ... we did not parsed it because the the map is  unordred and msg should be ordred
// get or create convertation ... used  return the id of the convertation
// for each user msg need to update the userstatue order......
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

// create a convertaion if not exist bewtweween two users ,, if exist it return its id
// int64 because it's long
// last insetreted id is func to bring the last id ...used with autoINcreamnt usually
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

// when convertation is already opened ....  no scroll .... send receive ... real_time_exchange
// bring the connection of the two users the sender and the receiver ....
// if both are online.... send if one online send (sender)
func deliverMessageToUser(senderID, receiverID, content string, conversationID int64) {
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
		"sent_at":         time.Now().Format("15:00"),
	}

	jsonMsg, err := json.Marshal(messagePayload)
	if err != nil {
		log.Println("Error marshaling message to deliver:", err)
		return
	}
	if receiverOnline {
		receiverConn.WriteMu.Lock()
		receiverConn.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
		receiverConn.WriteMu.Unlock()
	}
	if senderOnline {
		senderConn.WriteMu.Lock()
		senderConn.Conn.WriteMessage(websocket.TextMessage, jsonMsg)
		senderConn.WriteMu.Unlock()
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
		m.SentAt = sentTime.Format("15:04")
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
		m.SentAt = sentTime.Format("15:04")
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
