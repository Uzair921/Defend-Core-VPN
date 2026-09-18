package engine

import (
"net"
"sync"
"time"

"defendcore-vpn/internal/policy"
noisepkg "defendcore-vpn/internal/vpn/noise"
)

type Session struct {
ID         string
UserID     string
DeviceID   string
BackendID  string
AssignedIP net.IP
Noise      *noisepkg.Session
PeerAddr   *net.UDPAddr
Policy     *policy.Engine
CreatedAt  time.Time
LastSeen   time.Time
BytesIn    uint64
BytesOut   uint64
mu         sync.Mutex
}

func (s *Session) Touch() {
s.mu.Lock()
s.LastSeen = time.Now()
s.mu.Unlock()
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

func (s *Session) IdleFor() time.Duration {
s.mu.Lock()
defer s.mu.Unlock()
return time.Since(s.LastSeen)
}

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
IdleTimeout  time.Duration
}

func (c *Config) idleTimeout() time.Duration {
if c.IdleTimeout > 0 {
return c.IdleTimeout
}
return 3 * time.Minute
}
