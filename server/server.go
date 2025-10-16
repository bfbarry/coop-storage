package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	uploadDir = "./uploads"
	maxUploadSize = 10 << 20 // 10 MB
)

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatal("Failed to create uploads directory:", err)
	}

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello, world!")
	})
	port := ":8080"
	fmt.Printf("Server starting on port %s\n", port)
	fmt.Printf("Upload endpoint: http://localhost%s/upload\n", port)
	fmt.Printf("Download endpoint: http://localhost%s/download/{filename}\n", port)
	
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := filepath.Ext(header.Filename)
	if ext != ".jpg" && ext != ".jpeg" {
		http.Error(w, "Only JPEG files are allowed", http.StatusBadRequest)
		return
	}

	// Create destination file
	destPath := filepath.Join(uploadDir, header.Filename)
	dest, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create file on server", http.StatusInternalServerError)
		return
	}
	defer dest.Close()

	// Copy file contents
	if _, err := io.Copy(dest, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	log.Printf("File uploaded successfully: %s\n", header.Filename)
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully: %s\n", header.Filename)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract filename from URL path
	filename := filepath.Base(r.URL.Path)
	if filename == "." || filename == "/" {
		http.Error(w, "Filename not provided", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(uploadDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set headers
	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

	// Copy file to response
	if _, err := io.Copy(w, file); err != nil {
		log.Printf("Error sending file: %v\n", err)
		return
	}

	log.Printf("File downloaded: %s\n", filename)
}