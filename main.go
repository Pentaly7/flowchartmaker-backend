package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gookit/color"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// FlowchartData represents the flowchart content
type FlowchartData struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
type FileEntry struct {
	Key      string      `json:"key"`
	Type     string      `json:"type"` // "file" or "dir"
	Path     string      `json:"path"` // Relative path from storage root
	Children []FileEntry `json:"children,omitempty"`
}

// Config holds application configuration
type Config struct {
	StorageDir string
}

var config Config

func init() {
	// Set default storage directory to "projects" in current directory
	config.StorageDir = "storage"

	// Create storage directory if it doesn't exist
	if _, err := os.Stat(config.StorageDir); os.IsNotExist(err) {
		err := os.MkdirAll(config.StorageDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create storage directory: %v", err)
		}
	}
}

func main() {
	// Create server
	mux := http.NewServeMux()
	mux.HandleFunc("/flowcharts/", handleFlowcharts)

	handler := enableCORS(mux)
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: handler,
	}

	// Show ASCII art
	printBanner()

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\nShutting down server...")
}

func printBanner() {
	blue := color.FgLightBlue.Render
	cyan := color.FgCyan.Render
	green := color.FgGreen.Render
	yellow := color.FgYellow.Render

	asciiArt := `
███████╗███████╗██████╗ ██╗   ██╗███████╗██████╗ 
██╔════╝██╔════╝██╔══██╗██║   ██║██╔════╝██╔══██╗
███████╗█████╗  ██████╔╝██║   ██║█████╗  ██████╔╝
╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██╔══╝  ██╔══██╗
███████║███████╗██║  ██║ ╚████╔╝ ███████╗██║  ██║
╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝
	`
	formatWithNewline := "%s %s %s\n"
	fmt.Println(blue(asciiArt))
	fmt.Printf(formatWithNewline, cyan("➜"), yellow("Listening on:"), green("http://localhost:8080"))
	fmt.Printf(formatWithNewline, cyan("➜"), yellow("PID:"), green(os.Getpid()))
	fmt.Printf(formatWithNewline, cyan("➜"), yellow("Go Version:"), green(runtime.Version()))
	fmt.Printf(formatWithNewline, cyan("➜"), yellow("Started at:"), green(time.Now().Format("2006-01-02 15:04:05")))
	fmt.Println(blue("══════════════════════════════════════════════════"))
}

func handleFlowcharts(w http.ResponseWriter, r *http.Request) {
	// Extract and sanitize path
	path := strings.TrimPrefix(r.URL.Path, "/flowcharts/")
	safePath := filepath.Clean(filepath.Join(config.StorageDir, path))

	// SECURITY: Prevent directory traversal
	if !strings.HasPrefix(safePath, config.StorageDir) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		handleGet(w, safePath)
	case http.MethodPost, http.MethodPut:
		handleWrite(w, r, safePath)
	case http.MethodDelete:
		handleDelete(w, safePath)
	}
}

func handleGet(w http.ResponseWriter, path string) {
	info, err := os.Stat(path)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		entries, err := getDirectoryTree(path)
		if err != nil {
			http.Error(w, "Error reading directory", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
		return
	}

	// Handle file request
	content, _ := os.ReadFile(path)
	w.Write(content)
}

func getDirectoryTree(rootPath string) ([]FileEntry, error) {
	var entries []FileEntry

	// Get relative path from storage root
	relPath, _ := filepath.Rel(config.StorageDir, rootPath)
	relPath = path.Clean(filepath.ToSlash(relPath)) // Convert to forward slashes

	// Read directory contents
	dirEntries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		fullPath := filepath.Join(rootPath, entry.Name())
		childRelPath := path.Join(relPath, entry.Name())

		node := FileEntry{
			Key:  entry.Name(),
			Path: childRelPath,
			Type: "file",
		}

		if entry.IsDir() {
			node.Type = "dir"
			children, err := getDirectoryTree(fullPath)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}

		entries = append(entries, node)
	}

	return entries, nil
}

func handleWrite(w http.ResponseWriter, r *http.Request, path string) {

	decoder := json.NewDecoder(r.Body)
	var data FlowchartData
	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	pathWithExt := path + ".txt"
	// Create directory if it doesn't exist
	os.MkdirAll(filepath.Dir(pathWithExt), 0755)

	if data.Title != "" {
		os.WriteFile(pathWithExt, []byte(data.Content), 0644)
	}
	w.WriteHeader(http.StatusCreated)
}

func handleDelete(w http.ResponseWriter, path string) {
	// Delete file or empty directory
	os.RemoveAll(path)
	w.WriteHeader(http.StatusNoContent)
}

// CORS middleware
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins (change in production!)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass to next handler
		next.ServeHTTP(w, r)
	})
}
