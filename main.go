package main

import (
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"health-hub/internal/config"
	"health-hub/internal/handlers"
	"health-hub/internal/storage"
)

func main() {
	cfg := config.Load()

	// Initialize storage
	var store storage.Storage
	var err error
	
	if cfg.UseS3 && cfg.S3Bucket != "" {
		store, err = storage.NewS3Storage(cfg.DataPath, cfg.S3Bucket)
		if err != nil {
			log.Printf("Failed to initialize S3 storage: %v, falling back to file storage", err)
			store = storage.NewFileStorage(cfg.DataPath)
		} else {
			log.Println("Using S3 storage with local backup")
		}
	} else {
		store = storage.NewFileStorage(cfg.DataPath)
		log.Println("Using file storage")
	}

	// Initialize handlers
	h := handlers.NewHandlers(store)

	// Setup routes
	mux := http.NewServeMux()
	
	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	
	// API routes
	mux.HandleFunc("/", h.Home)
	mux.HandleFunc("/upload", h.Upload)
	mux.HandleFunc("/activities", h.Activities)
	mux.HandleFunc("/stats", h.Stats)
	mux.HandleFunc("/bulk-upload", h.BulkUpload)
	mux.HandleFunc("/activity/", h.ActivityDetail)
	mux.HandleFunc("/api/activities", h.GetActivities)
	mux.HandleFunc("/api/health", h.GetHealthMetrics)
	mux.HandleFunc("/api/upload/gpx", h.UploadGPX)
	mux.HandleFunc("/api/upload/health", h.UploadHealthData)
	mux.HandleFunc("/api/upload/bulk-gpx", h.BulkUploadGPX)
	mux.HandleFunc("/api/stats/activities", h.StatsActivities)
	mux.HandleFunc("/api/stats/health", h.StatsHealth)

	fmt.Printf("=== Health Hub Server ===\n")
	fmt.Printf("Starting server on port %s\n", cfg.Port)
	fmt.Printf("Local access: http://localhost:%s\n", cfg.Port)
	
	// Show Tailscale IP if available
	if err := showTailscaleInfo(cfg.Port); err != nil {
		log.Printf("Could not get Tailscale info: %v", err)
	}
	
	fmt.Printf("=== Server Running ===\n")
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}

func showTailscaleInfo(port string) error {
	// Try to get Tailscale IP
	cmd := exec.Command("tailscale", "ip", "-4")
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	
	ips := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(ips) > 0 && ips[0] != "" {
		fmt.Printf("\nTailscale access:\n")
		for _, ip := range ips {
			if ip != "" {
				fmt.Printf("  http://%s:%s\n", strings.TrimSpace(ip), port)
			}
		}
		fmt.Printf("\n")
	}
	
	return nil
}