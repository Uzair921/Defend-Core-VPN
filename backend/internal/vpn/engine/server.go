package engine

import (
"encoding/json"
"fmt"
"log"
"net"
"sync"
"time"

"defendcore-vpn/internal/policy"
noisepkg "defendcore-vpn/internal/vpn/noise"
"defendcore-vpn/internal/vpn/transport"
"defendcore-vpn/internal/vpn/tun"
)

type SessionHooks struct {
OnStart  func(userID, deviceID, clientIP, assignedIP string) (sessionID string, err error)
OnEnd    func(sessionID string, bytesIn, bytesOut int64) error
OnUpdate func(sessionID string, bytesIn, bytesOut int64) error
}

type Server struct {
cfg          Config
tunDev       tun.Device
udp          *transport.UDP
staticKP     *noisepkg.StaticKeypair
hooks        *SessionHooks
policyClient *PolicyClient
pool         *IPPool
sessions     map[string]*Session
byIP         map[string]*Session
mu           sync.RWMutex
}

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
cidr := cfg.AssignedCIDR
if cidr == "" {
cidr = "10.8.0.0/24"
}
pool, err := NewIPPool(cidr, true)
if err != nil {
tunDev.Close()
udpConn.Close()
return nil, fmt.Errorf("ip pool: %w", err)
}
return &Server{
cfg:      cfg,
tunDev:   tunDev,
udp:      udpConn,
staticKP: staticKP,
pool:     pool,
sessions: make(map[string]*Session),
byIP:     make(map[string]*Session),
}, nil
}

func (s *Server) SetPolicyClient(pc *PolicyClient) { s.policyClient = pc }
func (s *Server) SetHooks(h *SessionHooks)         { s.hooks = h }

func (s *Server) Run() error {
log.Printf("[server] TUN %s up (%s)", s.tunDev.Name(), s.cfg.TUNIP)
log.Printf("[server] listening on UDP :%d", s.cfg.ServerPort)
go s.udpLoop()
go s.tunLoop()
go s.cleanupLoop()
go s.statsLoop()
select {}
}

func (s *Server) Close() {
if s.tunDev != nil {
s.tunDev.Close()
}
if s.udp != nil {
s.udp.Close()
}
}

func (s *Server) udpLoop() {
buf := make([]byte, 65535)
for {
n, addr, err := s.udp.Recv(buf)
if err != nil {
log.Printf("[server] udp recv: %v", err)
continue
}
pkt := make([]byte, n)
copy(pkt, buf[:n])
s.mu.RLock()
sess := s.findSessionByAddr(addr)
s.mu.RUnlock()
if sess != nil {
s.handleData(sess, pkt, addr)
continue
}
s.handleHandshake(pkt, addr)
}
}

func (s *Server) handleData(sess *Session, ciphertext []byte, addr *net.UDPAddr) {
plaintext, err := sess.Noise.Decrypt(ciphertext)
if err != nil {
log.Printf("[server] decrypt from %s: %v", addr, err)
return
}
sess.AddIn(uint64(len(ciphertext)))
if !addrEqual(sess.PeerAddr, addr) {
s.mu.Lock()
sess.PeerAddr = addr
s.mu.Unlock()
}
if len(plaintext) < 20 {
return
}
if plaintext[0]>>4 != 4 {
return
}
if sess.Policy != nil {
dst := net.IP(plaintext[16:20]).String()
allowed, _ := sess.Policy.CheckDestination(dst, "")
if !allowed {
return
}
}
if _, err := s.tunDev.Write(plaintext); err != nil {
log.Printf("[server] tun write: %v", err)
}
}

func (s *Server) tunLoop() {
buf := make([]byte, 65535)
for {
n, err := s.tunDev.Read(buf)
if err != nil {
log.Printf("[server] tun read: %v", err)
continue
}
if n < 20 || buf[0]>>4 != 4 {
continue
}
dstIP := net.IP(buf[16:20]).String()
s.mu.RLock()
target := s.byIP[dstIP]
s.mu.RUnlock()
if target == nil {
continue
}
ciphertext, err := target.Noise.Encrypt(buf[:n])
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

type clientHello struct {
UserID   string `json:"user_id"`
DeviceID string `json:"device_id"`
}

func (s *Server) handleHandshake(msg1 []byte, addr *net.UDPAddr) {
cfg, err := noisepkg.ServerConfig(s.staticKP)
if err != nil {
log.Printf("[server] noise config: %v", err)
return
}
hs, payload, err := noisepkg.ServerStep1(cfg, msg1)
if err != nil {
log.Printf("[server] handshake step1 from %s: %v", addr, err)
return
}
var hello clientHello
if len(payload) > 0 {
_ = json.Unmarshal(payload, &hello)
}
sessionKey := hello.DeviceID
if sessionKey == "" {
sessionKey = addr.String()
}
assigned, err := s.pool.Allocate(sessionKey)
if err != nil {
log.Printf("[server] ip allocate: %v", err)
return
}
reply := map[string]string{
"status":      "ok",
"assigned_ip": assigned.String() + "/24",
}
replyBytes, _ := json.Marshal(reply)
noiseSess, msg2, err := noisepkg.ServerStep2(hs, replyBytes)
if err != nil {
s.pool.Release(assigned)
log.Printf("[server] handshake step2: %v", err)
return
}
if err := s.udp.SendTo(msg2, addr); err != nil {
s.pool.Release(assigned)
log.Printf("[server] send msg2: %v", err)
return
}
now := time.Now()
sess := &Session{
ID:         sessionKey,
UserID:     hello.UserID,
DeviceID:   hello.DeviceID,
AssignedIP: assigned,
Noise:      noiseSess,
PeerAddr:   addr,
CreatedAt:  now,
LastSeen:   now,
}
if s.hooks != nil && s.hooks.OnStart != nil {
id, err := s.hooks.OnStart(sess.UserID, sess.DeviceID, addr.IP.String(), assigned.String())
if err != nil {
log.Printf("[server] session start hook: %v", err)
} else {
sess.BackendID = id
}
}
s.mu.Lock()
if old, ok := s.sessions[sessionKey]; ok {
s.pool.Release(old.AssignedIP)
delete(s.byIP, old.AssignedIP.String())
}
s.sessions[sessionKey] = sess
s.byIP[assigned.String()] = sess
s.mu.Unlock()
log.Printf("[server] session established peer=%s device=%s ip=%s", addr, hello.DeviceID, assigned)
if s.policyClient != nil && sess.UserID != "" {
go s.loadPolicy(sess)
}
}

func (s *Server) loadPolicy(sess *Session) {
set, err := s.policyClient.FetchPolicy(sess.UserID, sess.DeviceID)
if err != nil {
log.Printf("[policy] fetch user=%s: %v (deny-all)", sess.UserID, err)
return
}
sess.Policy = policy.NewEngine(set.Rules)
log.Printf("[policy] loaded %d rules for user=%s", len(set.Rules), sess.UserID)
}

func (s *Server) cleanupLoop() {
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()
for range ticker.C {
timeout := s.cfg.idleTimeout()
var expired []*Session
s.mu.Lock()
for key, sess := range s.sessions {
if sess.IdleFor() > timeout {
expired = append(expired, sess)
delete(s.sessions, key)
delete(s.byIP, sess.AssignedIP.String())
s.pool.Release(sess.AssignedIP)
}
}
s.mu.Unlock()
for _, sess := range expired {
log.Printf("[server] session expired device=%s ip=%s", sess.DeviceID, sess.AssignedIP)
if s.hooks != nil && s.hooks.OnEnd != nil && sess.BackendID != "" {
in, out := sess.Stats()
_ = s.hooks.OnEnd(sess.BackendID, int64(in), int64(out))
}
}
}
}

func (s *Server) statsLoop() {
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()
for range ticker.C {
if s.hooks == nil || s.hooks.OnUpdate == nil {
continue
}
s.mu.RLock()
list := make([]*Session, 0, len(s.sessions))
for _, sess := range s.sessions {
list = append(list, sess)
}
s.mu.RUnlock()
for _, sess := range list {
if sess.BackendID == "" {
continue
}
in, out := sess.Stats()
_ = s.hooks.OnUpdate(sess.BackendID, int64(in), int64(out))
}
}
}

func (s *Server) findSessionByAddr(addr *net.UDPAddr) *Session {
for _, sess := range s.sessions {
if addrEqual(sess.PeerAddr, addr) {
return sess
}
}
return nil
}

func addrEqual(a, b *net.UDPAddr) bool {
if a == nil || b == nil {
return false
}
return a.IP.Equal(b.IP) && a.Port == b.Port
}
