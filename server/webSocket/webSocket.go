package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	g "forum/server/global"
	session "forum/server/session"

	"github.com/gorilla/websocket"
)

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

	// Add connection to active connections
	g.ActiveConnectionsMutex.Lock()
	g.ActiveConnections[userID] = &g.SafeConn{Conn: conn}
	g.ActiveConnectionsMutex.Unlock()

	// 🔄 Notify all clients: someone came online
	BroadcastUserStatus()

	// Clean up on disconnect
	defer func() {
		g.ActiveConnectionsMutex.Lock()
		delete(g.ActiveConnections, userID)
		g.ActiveConnectionsMutex.Unlock()

		// 🔄 Notify all clients: someone went offline
		BroadcastUserStatus()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		fmt.Printf("Received: %s\n", msg)
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
