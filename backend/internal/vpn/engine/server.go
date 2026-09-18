package engine

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"

	noisepkg "defendcore-vpn/internal/vpn/noise"
	"defendcore-vpn/internal/vpn/transport"
	"defendcore-vpn/internal/policy"
	"defendcore-vpn/internal/vpn/tun"
)

// Server is the VPN server engine.
// SessionHooks are callbacks the VPN server invokes on session lifecycle events.
type SessionHooks struct {
	OnStart  func(userID, deviceID, clientIP, assignedIP string) (sessionID string, err error)
	OnEnd    func(sessionID string, bytesIn, bytesOut int64) error
	OnUpdate func(sessionID string, bytesIn, bytesOut int64) error
}

type Server struct {
	cfg       Config
	tunDev    tun.Device
	udp       *transport.UDP
	staticKP  *noisepkg.StaticKeypair
	hooks     *SessionHooks
	policyClient *PolicyClient
	serverCfg interface{} // noise.Config
	sessions  map[string]*Session // key: peer addr string
	mu        sync.RWMutex
}

// NewServer creates a new VPN server.
func NewServer(cfg Config, staticKP *noisepkg.StaticKeypair) (*Server, error) {
	tunDev, err := tun.Open(cfg.TUNName)
	if err != nil {
		return nil, fmt.Errorf("open tun: %w", err)
	}
	if err := tunDev.Configure(cfg.TUNIP, cfg.TUNMask, cfg.TUNMTU); err != nil {
		tunDev.Close()
		return nil, fmt.Errorf("configure tun: %w", err)
	}

	udpConn, err := transport.Listen(fmt.Sprintf("0.0.0.0:%d", cfg.ServerPort))
	if err != nil {
		tunDev.Close()
		return nil, fmt.Errorf("listen udp: %w", err)
	}

	return &Server{
		cfg:      cfg,
		tunDev:   tunDev,
		udp:      udpConn,
		staticKP: staticKP,
		sessions: make(map[string]*Session),
	}, nil
}

// SetPolicyClient registers the policy client.
func (s *Server) SetPolicyClient(pc *PolicyClient) {
	s.policyClient = pc
}

// SetHooks registers session lifecycle callbacks.
func (s *Server) SetHooks(h *SessionHooks) {
	s.hooks = h
}

// Run starts the server loops.
func (s *Server) Run() error {
	log.Printf("[server] TUN %s up, IP %s", s.tunDev.Name(), s.cfg.TUNIP)
	log.Printf("[server] UDP listening on :%d", s.cfg.ServerPort)

	go s.udpLoop()
	go s.tunLoop()
	go s.cleanupLoop()
	go s.statsLoop()

	select {}
}

func (s *Server) udpLoop() {
	buf := make([]byte, 65535)
	for {
		n, addr, err := s.udp.Recv(buf)
		if err != nil {
			log.Printf("[server] udp recv: %v", err)
			continue
		}
		packet := make([]byte, n)
		copy(packet, buf[:n])

		s.mu.RLock()
		sess, ok := s.sessions[addr.String()]
		s.mu.RUnlock()

		if ok {
			// Encrypted data packet
			plaintext, err := sess.Noise.Decrypt(packet)
			if err != nil {
				log.Printf("[server] decrypt from %s: %v", addr, err)
				continue
			}
			sess.AddIn(uint64(n))

			// Debug: dump packet info
			if len(plaintext) < 20 {
				log.Printf("[server] short packet: %d bytes", len(plaintext))
				continue
			}
			version := plaintext[0] >> 4
			if version != 4 {
				log.Printf("[server] non-IPv4: version=%d first_byte=0x%02x len=%d",
					version, plaintext[0], len(plaintext))
				continue
			}
			proto := plaintext[9]
			srcIP := net.IP(plaintext[12:16])
			dstIP := net.IP(plaintext[16:20])
			log.Printf("[server] TUN write: %d bytes proto=%d src=%s dst=%s",
				len(plaintext), proto, srcIP, dstIP)

			if _, err := s.tunDev.Write(plaintext); err != nil {
				log.Printf("[server] tun write: %v", err)
			}
			continue
		}

		// New handshake — packet is msg1
		s.handleHandshake(packet, addr)
	}
}

func (s *Server) handleHandshake(msg1 []byte, addr *net.UDPAddr) {
	serverCfg, err := noisepkg.ServerConfig(s.staticKP)
	if err != nil {
		log.Printf("[server] config: %v", err)
		return
	}

	hs, payload, err := noisepkg.ServerStep1(serverCfg, msg1)
	if err != nil {
		log.Printf("[server] handshake step1 from %s: %v", addr, err)
		return
	}

	var clientInfo map[string]interface{}
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &clientInfo)
	}
	log.Printf("[server] handshake from %s, client info: %v", addr, clientInfo)

	session, msg2, err := noisepkg.ServerStep2(hs, []byte(`{"status":"ok"}`))
	if err != nil {
		log.Printf("[server] handshake step2: %v", err)
		return
	}

	log.Printf("[server] Sending msg2 (%d bytes) to %s", len(msg2), addr)
	if err := s.udp.SendTo(msg2, addr); err != nil {
		log.Printf("[server] send msg2 FAILED: %v", err)
	} else {
		log.Printf("[server] sent msg2 (%d bytes) to %s", len(msg2), addr)
	}

	assignedIP := stringVal(clientInfo, "ip")
	// Extract just the IP (strip /24 if present)
	if idx := strings.Index(assignedIP, "/"); idx > 0 {
		assignedIP = assignedIP[:idx]
	}

	newSess := &Session{
		Noise:      session,
		PeerAddr:   addr,
		CreatedAt:  time.Now(),
		LastSeen:   time.Now(),
		UserID:     stringVal(clientInfo, "user_id"),
		DeviceID:   stringVal(clientInfo, "device_id"),
		AssignedIP: assignedIP,
	}

	// Call backend hook
	if s.hooks != nil && s.hooks.OnStart != nil {
		assignedIP := stringVal(clientInfo, "ip")
		backendID, err := s.hooks.OnStart(
			newSess.UserID,
			newSess.DeviceID,
			addr.IP.String(),
			assignedIP,
		)
		if err != nil {
			log.Printf("[server] backend session start failed: %v", err)
		} else {
			newSess.BackendID = backendID
			log.Printf("[server] backend session ID: %s", backendID)
		}
	}

	s.mu.Lock()
	s.sessions[addr.String()] = newSess
	s.mu.Unlock()

	log.Printf("[server] session established with %s", addr)

	// Fetch access policy for this user+device (async, non-blocking)
	if s.policyClient != nil && newSess.UserID != "" {
		go s.applyPolicy(newSess)
	}

	// Fetch access policy for this user+device (async, non-blocking)
	if s.policyClient != nil && newSess.UserID != "" {
	}
}

func (s *Server) tunLoop() {
	buf := make([]byte, 65535)
	for {
		n, err := s.tunDev.Read(buf)
		if err != nil {
			log.Printf("[server] tun read ERROR: %v", err)
			continue
		}
		log.Printf("[server] TUN READ: %d bytes, first byte=0x%02x", n, buf[0])
		packet := make([]byte, n)
		copy(packet, buf[:n])

		// Parse IPv4 header to find destination
		if n < 20 || packet[0]>>4 != 4 {
			continue
		}
		dstIP := net.IP(packet[16:20]).String()

		// Find target session by AssignedIP
		var target *Session
		s.mu.RLock()
		for _, sess := range s.sessions {
			if sess.AssignedIP == dstIP {
				target = sess
				break
			}
		}
		s.mu.RUnlock()

		if target == nil {
			log.Printf("[server] no session for dst=%s, dropping", dstIP)
			continue
		}

		ciphertext, err := target.Noise.Encrypt(packet)
		if err != nil {
			log.Printf("[server] encrypt: %v", err)
			continue
		}
		if err := s.udp.SendTo(ciphertext, target.PeerAddr); err != nil {
			log.Printf("[server] send: %v", err)
			continue
		}
		target.AddOut(uint64(len(ciphertext)))
	}
}

func (s *Server) Close() {
	if s.tunDev != nil {
		s.tunDev.Close()
	}
	if s.udp != nil {
		s.udp.Close()
	}
}


func stringVal(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}


// cleanupLoop removes stale sessions and notifies the backend.
func (s *Server) cleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		var toEnd []*Session

		s.mu.Lock()
		for addr, sess := range s.sessions {
			if now.Sub(sess.CreatedAt) > 5*time.Minute && sess.BytesIn == 0 {
				// stale session
				toEnd = append(toEnd, sess)
				delete(s.sessions, addr)
			}
		}
		s.mu.Unlock()

		for _, sess := range toEnd {
			if s.hooks != nil && s.hooks.OnEnd != nil && sess.BackendID != "" {
				in, out := sess.Stats()
				_ = s.hooks.OnEnd(sess.BackendID, int64(in), int64(out))
			}
		}
	}
}

// statsLoop periodically pushes byte counters to the backend.
func (s *Server) statsLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if s.hooks == nil || s.hooks.OnUpdate == nil {
			continue
		}
		s.mu.RLock()
		sessions := make([]*Session, 0, len(s.sessions))
		for _, sess := range s.sessions {
			sessions = append(sessions, sess)
		}
		s.mu.RUnlock()

		for _, sess := range sessions {
			if sess.BackendID == "" {
				continue
			}
			in, out := sess.Stats()
			_ = s.hooks.OnUpdate(sess.BackendID, int64(in), int64(out))
		}
	}
}


// applyPolicy fetches and applies access policy for a session.
func (s *Server) applyPolicy(sess *Session) {
	userID := sess.UserID
	deviceID := sess.DeviceID
	clientIP := sess.AssignedIP

	log.Printf("[policy] fetching policy for user=%s device=%s ip=%s", userID, deviceID, clientIP)

	policySet, err := s.policyClient.FetchPolicy(userID, deviceID)
	if err != nil {
		log.Printf("[policy] fetch failed: %v — applying default DENY", err)
		sess.PolicyEngine = nil
		// Default deny: no rules means all traffic dropped
		return
	}

	log.Printf("[policy] got %d rules for %s", len(policySet.Rules), userID)

	// Build engine
	engine := policy.NewEngine(policySet.Rules)
	sess.PolicyEngine = engine

	// Generate iptables rules
	rules := engine.GenerateFirewallRules(clientIP)
	for _, rule := range rules {
		if strings.HasPrefix(rule, "#") {
			continue
		}
		cmd := exec.Command("sh", "-c", rule)
		if out, err := cmd.CombinedOutput(); err != nil {
			log.Printf("[policy] iptables error: %v — %s", err, string(out))
		} else {
			log.Printf("[policy] applied: %s", rule)
		}
	}

	log.Printf("[policy] applied %d rules for %s", len(rules), clientIP)
}
