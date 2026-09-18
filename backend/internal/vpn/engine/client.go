package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	noisepkg "defendcore-vpn/internal/vpn/noise"
	"defendcore-vpn/internal/vpn/transport"
	"defendcore-vpn/internal/vpn/tun"
)

// Client is the VPN client engine.
type Client struct {
	cfg      Config
	tunDev   tun.Device
	udp      *transport.UDP
	session  *Session
	keepalive time.Duration
}

// NewClient creates a new VPN client.
func NewClient(cfg Config) (*Client, error) {
	tunDev, err := tun.Open(cfg.TUNName)
	if err != nil {
		return nil, fmt.Errorf("open tun: %w", err)
	}
	if err := tunDev.Configure(cfg.TUNIP, cfg.TUNMask, cfg.TUNMTU); err != nil {
		tunDev.Close()
		return nil, fmt.Errorf("configure tun: %w", err)
	}

	udpConn, err := transport.Dial(fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort))
	if err != nil {
		tunDev.Close()
		return nil, fmt.Errorf("dial udp: %w", err)
	}

	return &Client{
		cfg:       cfg,
		tunDev:    tunDev,
		udp:       udpConn,
		keepalive: 25 * time.Second,
	}, nil
}

// Run performs the handshake and starts the client loops.
func (c *Client) Run() error {
	log.Printf("[client] TUN %s up, IP %s", c.tunDev.Name(), c.cfg.TUNIP)

	clientCfg, err := noisepkg.ClientConfig(c.cfg.ServerPubKey)
	if err != nil {
		return fmt.Errorf("client config: %w", err)
	}

	clientInfoMap := map[string]string{
		"device": "cli",
		"ip":     c.cfg.TUNIP,
	}
	if c.cfg.UserID != "" {
		clientInfoMap["user_id"] = c.cfg.UserID
	}
	if c.cfg.DeviceID != "" {
		clientInfoMap["device_id"] = c.cfg.DeviceID
	}
	clientInfo, _ := json.Marshal(clientInfoMap)
	msg1, hs, err := noisepkg.ClientStep1(clientCfg, clientInfo)
	if err != nil {
		return fmt.Errorf("client step1: %w", err)
	}

	if err := c.udp.Send(msg1); err != nil {
		return fmt.Errorf("send msg1: %w", err)
	}
	log.Printf("[client] sent handshake msg1 (%d bytes)", len(msg1))

	buf := make([]byte, 65535)
	c.udp.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, _, err := c.udp.Recv(buf)
	if err != nil {
		return fmt.Errorf("recv msg2: %w", err)
	}
	log.Printf("[client] received msg2 (%d bytes)", n)

	session, payload, err := noisepkg.ClientStep2(hs, buf[:n])
	if err != nil {
		return fmt.Errorf("client step2: %w", err)
	}
	log.Printf("[client] handshake complete, server said: %s", string(payload))

	c.session = &Session{
		Noise:     session,
		CreatedAt: time.Now(),
	}

	// Clear read deadline — otherwise udpLoop will timeout every 10s
	c.udp.SetReadDeadline(time.Time{})

	go c.udpLoop()
	go c.tunLoop()
	go c.keepaliveLoop()

	select {}
}

func (c *Client) udpLoop() {
	buf := make([]byte, 65535)
	for {
		n, _, err := c.udp.Recv(buf)
		if err != nil {
			log.Printf("[client] udp recv: %v", err)
			continue
		}
		plaintext, err := c.session.Noise.Decrypt(buf[:n])
		if err != nil {
			log.Printf("[client] decrypt: %v", err)
			continue
		}
		c.session.AddIn(uint64(n))
		if _, err := c.tunDev.Write(plaintext); err != nil {
			log.Printf("[client] tun write: %v", err)
		}
	}
}

func (c *Client) tunLoop() {
	buf := make([]byte, 65535)
	for {
		n, err := c.tunDev.Read(buf)
		if err != nil {
			log.Printf("[client] tun read: %v", err)
			continue
		}
		packet := make([]byte, n)
		copy(packet, buf[:n])

		ciphertext, err := c.session.Noise.Encrypt(packet)
		if err != nil {
			log.Printf("[client] encrypt: %v", err)
			continue
		}
		if err := c.udp.Send(ciphertext); err != nil {
			log.Printf("[client] send: %v", err)
			continue
		}
		c.session.AddOut(uint64(len(ciphertext)))
	}
}

func (c *Client) keepaliveLoop() {
	ticker := time.NewTicker(c.keepalive)
	defer ticker.Stop()
	for range ticker.C {
		empty, err := c.session.Noise.Encrypt(nil)
		if err != nil {
			continue
		}
		_ = c.udp.Send(empty)
	}
}

func (c *Client) Close() {
	if c.tunDev != nil {
		c.tunDev.Close()
	}
	if c.udp != nil {
		c.udp.Close()
	}
}
