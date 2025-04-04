package main

import (
	"fmt"
	"github.com/Pentaly7/flowchartmaker-backend/internal/app"
	"github.com/Pentaly7/flowchartmaker-backend/internal/models"
	"github.com/gookit/color"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var config models.Config

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
	// Show ASCII art
	printBanner()

	server := app.New(&config)
	// start server
	go server.Start()

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

	formatWithNewline := "%s %s %s\n"
	fmt.Printf(formatWithNewline, cyan("➜"), yellow("Listening on:"), green("http://localhost:8080"))
	fmt.Printf(formatWithNewline, cyan("➜"), yellow("PID:"), green(os.Getpid()))
	fmt.Printf(formatWithNewline, cyan("➜"), yellow("Go Version:"), green(runtime.Version()))
	fmt.Printf(formatWithNewline, cyan("➜"), yellow("Started at:"), green(time.Now().Format("2006-01-02 15:04:05")))
	fmt.Println(blue("══════════════════════════════════════════════════"))
}
