package handlers

import (
	"database/sql"
	"encoding/json"
	"strings"
	g "forum/server/global"
    "time"
	"log"
	http "net/http"
	"github.com/google/uuid"
	session "forum/server/session"

	_ "github.com/mattn/go-sqlite3"
)

// this function insert posts to DB
func insertPost(db *sql.DB, post g.Post) error {
	query := `
		INSERT INTO posts (id, title, content, category)
		VALUES (?, ?, ?, ?);
	`

	_, err := db.Exec(query, post.ID, post.Title, post.Content, post.Category)

	return err
}


// this function parse post form and insert it to DB
func Getcreatepost(w http.ResponseWriter, r *http.Request) {
    if err := r.ParseForm(); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error": "Unable to parse form data",
        })
		return
	}
	var post g.Post
	
	post.ID = uuid.New().String()
	post.Title = r.FormValue("title")
	post.Content = r.FormValue("content")
    categories := r.Form["categories[]"]
	if len(categories) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Please select at least one category",
		})
		return
	}
	// Join categories into a comma-separated string
	post.Category = strings.Join(categories, ", ")

	
	if err := insertPost(g.DB, post); err != nil {
		log.Println("Error inserting post:", err)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error": "Failed to create post",
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Post created successfully",
		"post":    post,
	})
}

// this function fetch all posts in DB
func Getposts(w http.ResponseWriter, r *http.Request) {
    _, ok := session.GetSessionUsername(r)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    rows, err := g.DB.Query("SELECT id, title, content, category FROM posts ORDER BY rowid DESC")
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var posts []g.Post
    for rows.Next() {
        var post g.Post
        if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category); err != nil {
            continue
        }
        posts = append(posts, post)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

// this function fetch a single post
func GetSinglePost(w http.ResponseWriter, r *http.Request) {
    // Extract the post ID from the URL manually
    pathParts := strings.Split(r.URL.Path, "/")
    if len(pathParts) < 4 || pathParts[3] == "" {
        http.Error(w, "Post ID not specified", http.StatusBadRequest)
        return
    }
    id := pathParts[3] // post ID

    var post g.Post
    err := g.DB.QueryRow("SELECT id, title, content, category FROM posts WHERE id = ?", id).
        Scan(&post.ID, &post.Title, &post.Content, &post.Category)

    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Post not found", http.StatusNotFound)
        } else {
            http.Error(w, "Database error", http.StatusInternalServerError)
        }
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(post)
}

// this function fetch all comments in a post
func Getcomments(w http.ResponseWriter, r *http.Request) {
    // Extract the post ID from the URL
    pathParts := strings.Split(r.URL.Path, "/")
    if len(pathParts) < 4 || pathParts[3] == "" {
        http.Error(w, "Post ID not specified", http.StatusBadRequest)
        return
    }
    postID := pathParts[3] // post ID

    rows, err := g.DB.Query("SELECT id, post_id, author, content, created_at FROM comments WHERE post_id = ?", postID)
    if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var comments []g.Comment
    for rows.Next() {
        var comment g.Comment
        if err := rows.Scan(&comment.ID, &comment.PostID, &comment.Author, &comment.Content, &comment.CreatedAt); err != nil {
            continue
        }
        comments = append(comments, comment)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}

// this function hanlde comments creation
func Getcreatecomment(w http.ResponseWriter, r *http.Request) {
    // Parse the form data
    if err := r.ParseForm(); err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error": "Unable to parse form data",
        })
        return
    }

    // Extract post ID from the URL 
    pathParts := strings.Split(r.URL.Path, "/")
    if len(pathParts) < 4 || pathParts[3] == "" {
        http.Error(w, "Post ID not specified", http.StatusBadRequest)
        return
    }
    postID := pathParts[3] // post ID

    var comment g.Comment
    comment.ID = uuid.New().String()
    comment.PostID = postID
    if _, ok := session.GetSessionUsername(r); !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    } else {
        comment.Author, _ = session.GetSessionUsername(r)
    }
    comment.Content = r.FormValue("comment")
    comment.CreatedAt = time.Now()

    query := `
        INSERT INTO comments (id, post_id, author, content, created_at)
        VALUES (?, ?, ?, ?, ?);
    `
    _, err := g.DB.Exec(query, comment.ID, comment.PostID, comment.Author, comment.Content, comment.CreatedAt)
    
    if err != nil {
        log.Println("Error inserting comment:", err)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "success": false,
            "error": "Failed to create comment",
        })
        return
    }

    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "message": "Comment created successfully",
        "comment": comment,
    })
}