package ws

import (
	"encoding/json"
	"log"
	"net/http"

	g "forum/server/global"
	"forum/server/session"

	"github.com/gorilla/websocket"
)

// upgeader varibale from websocket STRCUT to make http socket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// a function to handdle webSocket request /ws
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// if session exist && not expire continue
	userID, err := session.GetSessionUserID(r)
	if err != nil {
		http.Error(w, `{"status":401, "message":"You must be logged in"}`, http.StatusUnauthorized)
		return
	}
	// upgrade to webSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Add connection to map
	// lock before add connection and unlock after add connection
	g.ActiveConnectionsMutex.Lock()
	g.ActiveConnections[userID] = &g.SafeConn{Conn: conn}
	g.ActiveConnectionsMutex.Unlock()

	// nofify user after connection (online user.....)
	BroadcastUserStatus()

	defer func() {
		g.ActiveConnectionsMutex.Lock()
		delete(g.ActiveConnections, userID)
		g.ActiveConnectionsMutex.Unlock()
		BroadcastUserStatus() // Notify all clients (user offline)
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}
		log.Printf("Received from %s: %s\n", userID, msg)
	}
}

func GetOnlineUsers() (map[string]bool, map[string]string) {
	online := make(map[string]bool)
	all := make(map[string]string)

	rows, err := g.DB.Query("SELECT id, username FROM users")
	if err != nil {
		log.Println("Database error fetching users:", err)
		return nil, nil
	}
	defer rows.Close()

	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Println("Error scanning user:", err)
			continue
		}
		online[id] = false
		all[id] = name
	}

	g.ActiveConnectionsMutex.RLock()
	for id := range g.ActiveConnections {
		online[id] = true
	}
	g.ActiveConnectionsMutex.RUnlock()

	return online, all
}

func BroadcastUserStatus() {
	online, all := GetOnlineUsers()
	if online == nil || all == nil {
		return
	}

	var userList []map[string]interface{}
	for id, name := range all {
		userList = append(userList, map[string]interface{}{
			"id":       id,
			"username": name,
			"online":   online[id],
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
		log.Println("Error marshaling broadcast:", err)
		return
	}

	g.ActiveConnectionsMutex.RLock()
	defer g.ActiveConnectionsMutex.RUnlock()

	for userID, conn := range g.ActiveConnections {
		conn.WriteMu.Lock()
		err := conn.Conn.WriteMessage(websocket.TextMessage, jsonUpdate)
		conn.WriteMu.Unlock()

		if err != nil {
			log.Println("Error sending to user", userID, ":", err)
		}
	}
}
