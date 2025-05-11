package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Item struct {
	ID int `json:"id"`
}

func main() {
	// Initialize database connection
	var err error
	db, err = sql.Open("mysql", "uaxeq67kmmnss0ny:CtuLtYrtXMpfMeYvmydJ@tcp(b8yvleemcdkesovatlnm-mysql.services.clever-cloud.com:3306)/b8yvleemcdkesovatlnm")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.Handle("/", http.FileServer(http.Dir("./static")))
	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// Handle image upload
	http.HandleFunc("/upload", uploadHandler)

	// Serve image by ID
	http.HandleFunc("/image", imageHandling)

	// Set up routes with CORS middleware
	http.HandleFunc("/api", corsMiddleware(getItem))
	http.HandleFunc("/images", corsMiddleware(imageHandler))

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Failed to read uploaded file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read file data", http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("INSERT INTO images (name, data) VALUES (?, ?)")
	if err != nil {
		fmt.Println("Errro")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(header.Filename, imageData)
	if err != nil {
		http.Error(w, "Insert failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	fmt.Fprintf(w, "Image %s uploaded successfully!", header.Filename)
}

func imageHandling(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing image ID", http.StatusBadRequest)
		return
	}

	var data []byte
	err := db.QueryRow("SELECT data FROM images WHERE id = ?", id).Scan(&data)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	contentType := http.DetectContentType(data)
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// CORS middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// If it's a preflight OPTIONS request, return immediately
		if r.Method == "OPTIONS" {
			return
		}

		// Call the next handler
		next(w, r)
	}
}

func getItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT id FROM images")
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var name string
	var data []byte
	err = db.QueryRow("SELECT name, data FROM images WHERE id = ?", id).Scan(&name, &data)
	if err != nil {
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}

	// Detect content type based on file extension or use default
	contentType := http.DetectContentType(data)
	if contentType == "application/octet-stream" {
		// Fallback to JPEG if detection fails
		contentType = "image/jpeg"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}
