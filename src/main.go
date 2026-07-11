package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Retrieve environment variables with fallback options
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func main() {
	// 開発用に .env ファイルが存在すれば読み込む (本番環境などファイルが無い場合はエラーを無視して続行)
	_ = godotenv.Load()

	listenAddr := getEnv("LISTEN_ADDR", "10.0.0.2:8080")
	bcastAddr := getEnv("BCAST_ADDR", "255.255.255.255")
	targetMAC := getEnv("TARGET_MAC", "00:00:00:00:00:00")
	targetIP := getEnv("TARGET_IP", "192.168.40.51:22")
	targetUser := getEnv("TARGET_USER", "John Doe")
	sshKeyPath := getEnv("SSH_KEY_PATH", "/opt/pve-wol/id_ed25519")

	http.HandleFunc("/wake", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := SendWOL(targetMAC, bcastAddr); err != nil {
			log.Printf("WOL Error: %v", err)
			http.Error(w, "Failed to send WOL", http.StatusInternalServerError)
			return
		}
		log.Println("WOL Magic Packet sent successfully")
		fmt.Fprintf(w, "OK\n")
	})

	http.HandleFunc("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := ShutdownPC(targetIP, targetUser, sshKeyPath); err != nil {
			log.Printf("Shutdown Error: %v", err)
			http.Error(w, "Failed to shutdown PC", http.StatusInternalServerError)
			return
		}
		log.Println("Shutdown command sent successfully")
		fmt.Fprintf(w, "OK\n")
	})

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := CheckStatus(targetIP); err != nil {
			log.Printf("Status Check: PC is offline (%v)", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "OFFLINE\n")
			return
		}
		log.Println("Status Check: PC is online")
		fmt.Fprintf(w, "ONLINE\n")
	})

	log.Printf("Starting secure WOL/Shutdown API on %s", listenAddr)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
