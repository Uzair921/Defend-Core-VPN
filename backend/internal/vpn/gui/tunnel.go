package gui

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"defendcore-vpn/internal/vpn/engine"
	"defendcore-vpn/internal/devices"
)

type Tunnel struct {
	client *engine.Client
	stats  *Stats
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	active bool
}

func NewTunnel(stats *Stats) *Tunnel {
	return &Tunnel{
		stats: stats,
	}
}

func (t *Tunnel) Start(serverHost string, serverPort int, serverPubKey, userID, deviceID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.active {
		return fmt.Errorf("tunnel already active")
	}

	pubKeyBytes, err := devices.DecodeKey(serverPubKey)
	if err != nil {
		return fmt.Errorf("decode server public key: %w", err)
	}

	cfg := engine.Config{
		ServerHost:   serverHost,
		ServerPort:   serverPort,
		TUNName:      "dcvpn0",
		TUNIP:        "10.8.0.2/24",
		TUNMask:      "255.255.255.0",
		TUNMTU:       1420,
		ServerPubKey: pubKeyBytes,
		ClientMode:   true,
		UserID:       userID,
		DeviceID:     deviceID,
	}

	client, err := engine.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	t.client = client
	t.ctx, t.cancel = context.WithCancel(context.Background())

	go func() {
		if err := client.Run(); err != nil {
			log.Printf("[tunnel] client run: %v", err)
		}
	}()

	t.active = true
	return nil
}

func (t *Tunnel) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.active {
		return
	}

	if t.cancel != nil {
		t.cancel()
	}
	if t.client != nil {
		t.client.Close()
	}
	t.active = false
}

func (t *Tunnel) IsActive() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.active
}

func (t *Tunnel) WaitForConnection(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if t.IsActive() {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("connection timeout")
}
