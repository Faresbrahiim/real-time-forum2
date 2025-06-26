package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	g "forum/server/global"
)

// CreateSession generates a unique session ID.
func CreateSession(userID string) string {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		log.Fatal("Failed to generate session ID:", err)
	}
	return hex.EncodeToString(id)
}

// SetSession creates and sets the session cookie, stores in memory, and returns ID + expiration for DB
func SetSession(w http.ResponseWriter, userID string) (string, time.Time) {
	sessionId := CreateSession(userID)
	expiration := time.Now().Add(24 * time.Hour)

	// Store in memory
	g.SessionsMu.Lock()
	g.Sessions[sessionId] = userID
	g.SessionsMu.Unlock()

	// Set cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		HttpOnly: true,
		Path:     "/",
		Expires:  expiration,
	}
	http.SetCookie(w, cookie)

	return sessionId, expiration
}

// GetSessionUsername returns the user ID from in-memory session store
func GetSessionUsername(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", false
	}

	g.SessionsMu.Lock()
	userID, exists := g.Sessions[cookie.Value]
	g.SessionsMu.Unlock()

	return userID, exists
}

// DeleteSession removes the session from memory and clears the cookie
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	log.Println("Deleting session")

	cookie, err := r.Cookie("session_id")
	if err != nil {
		return
	}

	// Remove from memory
	g.SessionsMu.Lock()
	delete(g.Sessions, cookie.Value)
	g.SessionsMu.Unlock()

	// Invalidate cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

// GetSessionUserID checks the session from DB and returns the user ID if valid
func GetSessionUserID(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		fmt.Println("Session cookie missing")
		return "", fmt.Errorf("missing cookie: %w", err)
	}

	fmt.Println("Session ID:", cookie.Value)

	var userID string
	var expiration time.Time
	err = g.DB.QueryRow("SELECT user_id, expires_at FROM Session WHERE id = ?", cookie.Value).Scan(&userID, &expiration)
	if err != nil {
		fmt.Println("DB query failed for session:", err)
		return "", fmt.Errorf("invalid session: %w", err)
	}

	if time.Now().After(expiration) {
		fmt.Println("Session expired")
		return "", fmt.Errorf("session expired")
	}

	fmt.Println("Session valid for userID:", userID)
	return userID, nil
}
