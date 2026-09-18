package engine

import (
	"net"
	"sync"
	"time"

	noisepkg "defendcore-vpn/internal/vpn/noise"
	"defendcore-vpn/internal/policy"
)

// Session represents an established Noise session with a peer.
type Session struct {
	Noise     *noisepkg.Session
	PeerAddr  *net.UDPAddr
	CreatedAt time.Time
	LastSeen  time.Time
	BytesIn   uint64
	BytesOut  uint64
	BackendID string // session ID from backend
	UserID    string
	DeviceID  string
	AssignedIP string // client's tunnel IP (e.g. 10.8.0.2)
	PolicyEngine *policy.Engine // access policy engine
	mu        sync.Mutex
}

func (s *Session) AddIn(n uint64) {
	s.mu.Lock()
	s.BytesIn += n
	s.LastSeen = time.Now()
	s.mu.Unlock()
}

func (s *Session) AddOut(n uint64) {
	s.mu.Lock()
	s.BytesOut += n
	s.mu.Unlock()
}

func (s *Session) Stats() (in, out uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.BytesIn, s.BytesOut
}

// StatsReporter allows callbacks when stats change.
type StatsReporter func(s *Session)

// Config holds engine configuration.
type Config struct {
	ServerHost   string
	ServerPort   int
	TUNName      string
	TUNIP        string
	TUNMask      string
	TUNMTU       int
	AssignedCIDR string
	ServerPubKey []byte
	ServerPriv   []byte
	ClientMode   bool
	UserID       string
	DeviceID     string
}
