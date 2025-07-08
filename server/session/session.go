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
func SetSession(w http.ResponseWriter, userID string, username string) (error) {
	sessionId := CreateSession(userID)
	expiration := time.Now().Add(24 * time.Hour)

	// Store in memory
	_, err := g.DB.Exec("INSERT INTO Session (id, user_id, username, expires_at) VALUES (?, ?, ?, ?)", sessionId, userID, username, expiration)
	if err != nil {
		log.Println("Failed to store session in DB:", err)
		// Still proceed so the client gets a response
	}
	// Set cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionId,
		HttpOnly: true,
		Path:     "/",
		Expires:  expiration,
	}
	http.SetCookie(w, cookie)
	return err
}


// GetSessionUsername returns the user ID from in-memory session store
func GetSessionUsername(r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", false
	}
	var username string
	var expiresAt time.Time
	exists := false
	err = g.DB.QueryRow("SELECT username, expires_at FROM Session WHERE id = ?", cookie.Value).
		Scan(&username, &expiresAt)
	if err != nil {
		return "", false
	}
	if time.Now().After(expiresAt) {
		g.DB.Exec("DELETE FROM Session WHERE id = ?", cookie.Value)
		return "", false
	}
	if (username != "") {
		exists = true
	}
	return username, exists
}

// DeleteSession removes the session from memory and clears the cookie
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	log.Println("Deleting session")

	cookie, err := r.Cookie("session_id")
	if err != nil {
		return
	}

	// Remove from memory
	_, err = g.DB.Exec("DELETE FROM Session WHERE id = ?", cookie.Value)
	if err != nil {
		log.Println("Failed to delete session from DB:", err)
	}

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
