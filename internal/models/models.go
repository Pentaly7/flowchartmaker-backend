package models

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
