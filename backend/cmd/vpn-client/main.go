package main

import (
	"encoding/base64"
	"log"
	"os"

	"defendcore-vpn/internal/vpn/engine"
)

func main() {
	serverHost := getenv("VPN_SERVER_HOST", "127.0.0.1")
	serverPub := os.Getenv("VPN_SERVER_PUBLIC_KEY")
	userID := os.Getenv("VPN_USER_ID")
	deviceID := os.Getenv("VPN_DEVICE_ID")
	if serverPub == "" {
		log.Fatal("VPN_SERVER_PUBLIC_KEY must be set")
	}
	serverPubBytes, err := base64.StdEncoding.DecodeString(serverPub)
	if err != nil {
		log.Fatalf("decode server pub: %v", err)
	}

		cfg := engine.Config{
		ServerHost:   serverHost,
		ServerPort:   51820,
		TUNName:      "dcvpn0",
		TUNIP:        "10.8.0.2/24",
		TUNMask:      "255.255.255.0",
		TUNMTU:       1420,
		ServerPubKey: serverPubBytes,
		ClientMode:   true,
		UserID:       userID,
		DeviceID:     deviceID,
	}

	cli, err := engine.NewClient(cfg)
	if err != nil {
		log.Fatalf("new client: %v", err)
	}
	defer cli.Close()

	log.Println("VPN client starting...")
	if err := cli.Run(); err != nil {
		log.Fatalf("client run: %v", err)
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
