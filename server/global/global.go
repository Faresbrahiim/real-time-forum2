package glo

import (
	"database/sql"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var DB *sql.DB

var (
	Sessions   = map[string]string{}
	SessionsMu sync.Mutex
)

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
// ===
type SafeConn struct {
	Conn    *websocket.Conn
	WriteMu sync.Mutex
}

var ActiveConnections = make(map[string]*SafeConn)
var ActiveConnectionsMutex = sync.RWMutex{}