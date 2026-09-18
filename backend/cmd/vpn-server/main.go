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

var serverID string

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

backendURL := envOr("BACKEND_URL", "http://127.0.0.1:8080")
apiKey := os.Getenv("VPN_API_KEY")
serverName := envOr("VPN_SERVER_NAME", "defendcore-01")
publicIP := envOr("VPN_PUBLIC_IP", "0.0.0.0")
cidr := envOr("VPN_SUBNET", "10.8.0.0/24")
tunIP := envOr("VPN_TUN_IP", "10.8.0.1/24")

cfg := engine.Config{
ServerPort:   51820,
TUNName:      envOr("VPN_INTERFACE", "dcvpn0"),
TUNIP:        tunIP,
TUNMask:      "255.255.255.0",
TUNMTU:       1420,
AssignedCIDR: cidr,
IdleTimeout:  3 * time.Minute,
}

srv, err := engine.NewServer(cfg, staticKP)
if err != nil {
log.Fatalf("new server: %v", err)
}
defer srv.Close()

srv.SetPolicyClient(engine.NewPolicyClient(backendURL, apiKey))
srv.SetHooks(makeHooks(backendURL, apiKey, &serverID))

go registerAndHeartbeat(backendURL, apiKey, serverName, publicIP, &serverID)

log.Println("VPN server starting")
if err := srv.Run(); err != nil {
log.Fatalf("server: %v", err)
}
}

func makeHooks(backendURL, apiKey string, sid *string) *engine.SessionHooks {
return &engine.SessionHooks{
OnStart: func(userID, deviceID, clientIP, assignedIP string) (string, error) {
body := map[string]interface{}{
"user_id":     userID,
"device_id":   deviceID,
"server_id":   *sid,
"client_ip":   clientIP,
"assigned_ip": assignedIP,
}
var out struct {
ID string `json:"id"`
}
if err := backendPost(backendURL, apiKey, "/api/v1/vpn/sessions", body, &out); err != nil {
return "", err
}
return out.ID, nil
},
OnUpdate: func(sessionID string, bytesIn, bytesOut int64) error {
body := map[string]int64{"bytes_in": bytesIn, "bytes_out": bytesOut}
return backendPatch(backendURL, apiKey, "/api/v1/vpn/sessions/"+sessionID, body, nil)
},
OnEnd: func(sessionID string, bytesIn, bytesOut int64) error {
body := map[string]int64{"bytes_in": bytesIn, "bytes_out": bytesOut}
return backendPost(backendURL, apiKey, "/api/v1/vpn/sessions/"+sessionID+"/end", body, nil)
},
}
}

func registerAndHeartbeat(backendURL, apiKey, name, publicIP string, sid *string) {
register := func() {
body := map[string]string{
"name":      name,
"public_ip": publicIP,
"region":    envOr("VPN_REGION", "default"),
"version":   "0.2.0",
}
var out map[string]interface{}
if err := backendPost(backendURL, apiKey, "/api/v1/vpn/servers/register", body, &out); err != nil {
log.Printf("[backend] register: %v", err)
return
}
if id, ok := out["id"].(string); ok {
*sid = id
log.Printf("[backend] registered as %s", id)
}
}
register()
ticker := time.NewTicker(30 * time.Second)
for range ticker.C {
if *sid == "" {
register()
continue
}
body := map[string]string{"server_id": *sid, "status": "online"}
if err := backendPost(backendURL, apiKey, "/api/v1/vpn/servers/heartbeat", body, nil); err != nil {
log.Printf("[backend] heartbeat: %v", err)
}
}
}

func backendPost(base, key, path string, body, out interface{}) error {
return backendDo(http.MethodPost, base, key, path, body, out)
}

func backendPatch(base, key, path string, body, out interface{}) error {
return backendDo(http.MethodPatch, base, key, path, body, out)
}

func backendDo(method, base, key, path string, body, out interface{}) error {
b, _ := json.Marshal(body)
req, err := http.NewRequest(method, base+path, bytes.NewReader(b))
if err != nil {
return err
}
req.Header.Set("Content-Type", "application/json")
if key != "" {
req.Header.Set("X-API-Key", key)
}
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

func loadEnv() {
candidates := []string{".env", "../.env", "../../.env"}
if exe, err := os.Executable(); err == nil {
dir := filepath.Dir(exe)
candidates = append(candidates, filepath.Join(dir, ".env"), filepath.Join(dir, "..", ".env"))
}
for _, p := range candidates {
if _, err := os.Stat(p); err == nil {
_ = godotenv.Load(p)
return
}
}
}

func envOr(key, fallback string) string {
if v := os.Getenv(key); v != "" {
return v
}
return fallback
}
