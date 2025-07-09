package glo

import (
	"database/sql"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var DB *sql.DB

type User struct {
	ID        string
	Username  string
	Email     string
	Age       int
	Gender    string
	FirstName string
	LastName  string
	Password  []byte
}

type Post struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
}

type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// // safeConn ,,, store the current connect  + the mutex to avoid two go routines makes action at same time
// // Every individual connection needs its own lock.
// // protect the map itself used for adding removing listing ,,,,.. action happend affect connection
// type SafeConn struct {
// 	Conn    *websocket.Conn
// 	WriteMu sync.Mutex
// }

// // active connection map that have in key the user id ,,,  and in valur a pointer to safeConeection struct ... the conn + mutex
// // protect the websockt connection from conncurent writes
// // Even if the map is safe, the WebSocket itself (Conn) is not safe for concurrent writes.
// // so it protect the map....
// // accept lock unlock rlock runlock  ,,, rlock just read not write ...
// var (
// 	ActiveConnections      = make(map[string]*SafeConn)
// 	ActiveConnectionsMutex = sync.RWMutex{}
// )

type SafeConn struct {
	Conn    *websocket.Conn
	WriteMu sync.Mutex
}

var ActiveConnections = make(map[string][]*SafeConn)
var ActiveConnectionsMutex = sync.RWMutex{}
