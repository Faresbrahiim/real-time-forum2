package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	g "forum/server/global"
	h "forum/server/handlers"
	ws "forum/server/webSocket"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	var err error
	g.DB, err = sql.Open("sqlite3", "file=../../server/database/database.db?_busy_timeout=2000&_journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}
	filePath := os.Getenv("MODULES_SQL_PATH")
	if filePath == "" {
		filePath = "./server/database/database.sql"
	}

	query, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}

	_, err = g.DB.Exec(string(query))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database migrated successfully")
}

func main() {
	tmpl, err := template.ParseFiles(filepath.Join("client", "templates", "index.html"))
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	fs := http.FileServer(http.Dir(filepath.Join("client", "static")))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// autentification handlers
	http.HandleFunc("/api/login", h.Getlogin)
	http.HandleFunc("/api/signup", h.Getregister)
	http.HandleFunc("/api/logout", h.Get)

	// posts handlers
	http.HandleFunc("/api/checksession", h.CheckSession)
	http.HandleFunc("/api/createpost", h.Getcreatepost)
	http.HandleFunc("/api/posts", h.Getposts)
	http.HandleFunc("/api/singlepost/", h.GetSinglePost)
	http.HandleFunc("/api/createcomment/", h.Getcreatecomment)
	http.HandleFunc("/api/comments/", h.Getcomments)


	// ws handllers
	http.HandleFunc("/api/messages", ws.GetMessagesHandler)
	http.HandleFunc("/api/latest-messages", ws.GetLatestMessagesHandler)
	http.HandleFunc("/ws", ws.HandleWebSocket)



	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			log.Println("Template execution error:", err)
		}
	})

	log.Println("Server running at http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
