package app

import (
	"encoding/json"
	"fmt"
	"github.com/Pentaly7/flowchartmaker-backend/internal/models"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func (s *httpServer) handleFlowcharts(w http.ResponseWriter, r *http.Request) {
	// Extract and sanitize path
	pathString := strings.TrimPrefix(r.URL.Path, "/flowcharts/")
	safePath := filepath.Clean(filepath.Join(s.cfg.StorageDir, pathString))

	// SECURITY: Prevent directory traversal
	if !strings.HasPrefix(safePath, s.cfg.StorageDir) {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, safePath)
	case http.MethodPost, http.MethodPut:
		s.handleWrite(w, r, safePath)
	case http.MethodDelete:
		s.handleDelete(w, safePath)
	}
}

func (s *httpServer) handleGet(w http.ResponseWriter, path string) {
	info, err := os.Stat(path)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	if info.IsDir() {
		entries, err := s.getDirectoryTree(path)
		if err != nil {
			http.Error(w, "Error reading directory", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(entries)
		if err != nil {
			fmt.Println(err)
			return
		}
		return
	}

	// Handle file request
	content, _ := os.ReadFile(path)
	_, err = w.Write(content)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *httpServer) getDirectoryTree(rootPath string) ([]models.FileEntry, error) {
	var entries []models.FileEntry

	// Get relative path from storage root
	relPath, _ := filepath.Rel(s.cfg.StorageDir, rootPath)
	relPath = path.Clean(filepath.ToSlash(relPath)) // Convert to forward slashes

	// Read directory contents
	dirEntries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range dirEntries {
		fullPath := filepath.Join(rootPath, entry.Name())
		childRelPath := path.Join(relPath, entry.Name())

		node := models.FileEntry{
			Key:  entry.Name(),
			Path: childRelPath,
			Type: "file",
		}

		if entry.IsDir() {
			node.Type = "dir"
			children, err := s.getDirectoryTree(fullPath)
			if err != nil {
				return nil, err
			}
			node.Children = children
		}

		entries = append(entries, node)
	}

	return entries, nil
}

func (s *httpServer) handleWrite(w http.ResponseWriter, r *http.Request, path string) {

	decoder := json.NewDecoder(r.Body)
	var data models.FlowchartData
	err := decoder.Decode(&data)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	pathWithExt := path + ".txt"
	// Create directory if it doesn't exist
	var dir string
	if data.Title != "" {
		dir = filepath.Dir(path)
	} else {
		dir = path
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return
	}

	if data.Title != "" {
		err := os.WriteFile(pathWithExt, []byte(data.Content), 0644)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
	w.WriteHeader(http.StatusCreated)
}

func (s *httpServer) handleDelete(w http.ResponseWriter, path string) {
	// Delete file or empty directory
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
