package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	http "net/http"
	"strconv"
	"strings"

	g "forum/server/global"
	session "forum/server/session"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func insertUser(db *sql.DB, user g.User) error {
	query := `
        INSERT INTO users (
            id, username, email, age, gender, firstName, lastName, password_hash
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
    `

	_, err := db.Exec(query, user.ID, user.Username, user.Email, user.Age, user.Gender, user.FirstName, user.LastName, user.Password)

	return err
}

func Getregister(w http.ResponseWriter, r *http.Request) {
	// Regular form parsing works now
	if err := r.ParseForm(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unable to parse form data",
		})
		return
	}

	var user g.User

	user.ID = uuid.New().String()
	user.Username = r.FormValue("Nickname")
	user.Email = r.FormValue("E-mail")

	age, err := strconv.Atoi(r.FormValue("Age"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid age value",
		})
		return
	}
	user.Age = age

	user.Gender = r.FormValue("gender")
	user.FirstName = r.FormValue("First Name")
	user.LastName = r.FormValue("Last Name")

	password := r.FormValue("Password")
	user.Password, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Error processing password",
		})
		return
	}

	if err := insertUser(g.DB, user); err != nil {
		log.Println("Insert error:", err)

		w.Header().Set("Content-Type", "application/json")

		errorMsg := "Registration failed. Please try again."
		errorStr := err.Error()

		if strings.Contains(errorStr, "UNIQUE constraint failed: users.email") {
			errorMsg = "Email already exists. Please use a different email."
		} else if strings.Contains(errorStr, "UNIQUE constraint failed: users.username") {
			errorMsg = "Username already exists. Please choose a different username."
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   errorMsg,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Registration successful!",
	})
}

func Getlogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := r.ParseForm(); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Unable to parse form data",
		})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Username and password are required",
		})
		return
	}

	log.Println("Login attempt for user:", username)
	log.Println("Password provided:", password)

	var hashedPassword string
	err := g.DB.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Account not found",
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Server error",
			})
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Incorrect password",
		})
		return
	}

	// Fetch user ID
	var userID string
	err = g.DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
	if err != nil {
		log.Println("Failed to fetch user ID:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Server error",
		})
		return
	}

	// Set session and store in DB
	sessionId, expiration := session.SetSession(w, userID)
	_, err = g.DB.Exec("INSERT INTO Session (id, user_id, expires_at) VALUES (?, ?, ?)", sessionId, userID, expiration)
	if err != nil {
		log.Println("Failed to store session in DB:", err)
		// Still proceed so the client gets a response
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Login successful",
	})
}

func CheckSession(w http.ResponseWriter, r *http.Request) {
	username, ok := session.GetSessionUsername(r)
	w.Header().Set("Content-Type", "application/json")

	if ok {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"loggedIn": true,
			"username": username,
		})
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"loggedIn": false,
		})
	}
}

func Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	session.DeleteSession(w, r)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": " successful",
	})
}
