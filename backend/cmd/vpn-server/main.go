package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"

	"defendcore-vpn/internal/vpn/engine"
	noisepkg "defendcore-vpn/internal/vpn/noise"
)

var (
	backendURL string
	apiKey     string
	serverID   string
	serverName string
	publicIP   string
)

func loadEnv() {
	candidates := []string{
		".env", "../.env", "../../.env",
		"/root/defendcore-vpn/.env",
		"/home/ubuntu/defendcore-vpn/.env",
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "..", ".env"),
			filepath.Join(exeDir, "..", "..", ".env"),
		)
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err == nil {
				log.Printf("[startup] loaded env from %s", p)
				return
			}
		}
	}
}

func backendPost(path string, body interface{}, out interface{}) error {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest("POST", backendURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func registerWithBackend() {
	body := map[string]string{
		"name":     serverName,
		"public_ip": publicIP,
		"region":   "lab",
		"version":  "0.1.0",
	}
	var out map[string]interface{}
	if err := backendPost("/api/v1/vpn/servers/register", body, &out); err != nil {
		log.Printf("[backend] register failed: %v", err)
		return
	}
	if id, ok := out["id"].(string); ok {
		serverID = id
		log.Printf("[backend] registered as server ID %s", serverID)
	}
}

func heartbeatLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if serverID == "" {
			registerWithBackend()
			continue
		}
		body := map[string]string{
			"server_id": serverID,
			"status":    "online",
		}
		if err := backendPost("/api/v1/vpn/servers/heartbeat", body, nil); err != nil {
			log.Printf("[backend] heartbeat failed: %v", err)
		}
	}
}

func main() {
	loadEnv()

	privB64 := os.Getenv("VPN_SERVER_PRIVATE_KEY")
	pubB64 := os.Getenv("VPN_SERVER_PUBLIC_KEY")
	if privB64 == "" || pubB64 == "" {
		log.Fatal("VPN_SERVER_PRIVATE_KEY and VPN_SERVER_PUBLIC_KEY must be set")
	}
	if _, err := base64.StdEncoding.DecodeString(privB64); err != nil {
		log.Fatalf("decode private key: %v", err)
	}

	staticKP, err := noisepkg.DecodeStaticKeypair(privB64, pubB64)
	if err != nil {
		log.Fatalf("decode keypair: %v", err)
	}

	backendURL = os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://127.0.0.1:8080"
	}
	apiKey = os.Getenv("VPN_API_KEY")
	serverName = os.Getenv("VPN_SERVER_NAME")
	if serverName == "" {
		serverName = "defendcore-lab-01"
	}
	publicIP = os.Getenv("VPN_PUBLIC_IP")
	if publicIP == "" {
		publicIP = "192.168.174.132"
	}

	cfg := engine.Config{
		ServerPort: 51820,
		TUNName:    "dcvpn0",
		TUNIP:      "10.8.0.1/24",
		TUNMask:    "255.255.255.0",
		TUNMTU:     1420,
	}

	srv, err := engine.NewServer(cfg, staticKP)
	if err != nil {
		log.Fatalf("new server: %v", err)
	}
	defer srv.Close()

	policyClient := engine.NewPolicyClient(backendURL, apiKey)
	srv.SetPolicyClient(policyClient)
	srv.SetHooks(makeHooks())

	// Register with backend
	registerWithBackend()
	go heartbeatLoop()

	log.Println("VPN server starting...")
	if err := srv.Run(); err != nil {
		log.Fatalf("server run: %v", err)
	}
}


// makeHooks returns SessionHooks that call the backend.
func makeHooks() *engine.SessionHooks {
	return &engine.SessionHooks{
		OnStart: func(userID, deviceID, clientIP, assignedIP string) (string, error) {
			body := map[string]interface{}{
				"user_id":     userID,
				"device_id":   deviceID,
				"server_id":   serverID,
				"client_ip":   clientIP,
				"assigned_ip": assignedIP,
			}
			var out struct {
				ID string `json:"id"`
			}
			if err := backendPost("/api/v1/vpn/sessions", body, &out); err != nil {
				return "", err
			}
			return out.ID, nil
		},
		OnUpdate: func(sessionID string, bytesIn, bytesOut int64) error {
			body := map[string]int64{
				"bytes_in":  bytesIn,
				"bytes_out": bytesOut,
			}
			return backendPatch("/api/v1/vpn/sessions/"+sessionID, body, nil)
		},
		OnEnd: func(sessionID string, bytesIn, bytesOut int64) error {
			body := map[string]int64{
				"bytes_in":  bytesIn,
				"bytes_out": bytesOut,
			}
			return backendPost("/api/v1/vpn/sessions/"+sessionID+"/end", body, nil)
		},
	}
}


func backendPatch(path string, body interface{}, out interface{}) error {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest("PATCH", backendURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}
