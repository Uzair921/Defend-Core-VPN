package engine

import (
"encoding/json"
"fmt"
"log"
"net"
"time"

noisepkg "defendcore-vpn/internal/vpn/noise"
"defendcore-vpn/internal/vpn/transport"
"defendcore-vpn/internal/vpn/tun"
)

type Client struct {
cfg       Config
tunDev    tun.Device
udp       *transport.UDP
session   *Session
keepalive time.Duration
}

func NewClient(cfg Config) (*Client, error) {
tunDev, err := tun.Open(cfg.TUNName)
if err != nil {
return nil, fmt.Errorf("open tun: %w", err)
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

func (c *Client) Run() error {
clientCfg, err := noisepkg.ClientConfig(c.cfg.ServerPubKey)
if err != nil {
return fmt.Errorf("client noise config: %w", err)
}
hello := map[string]string{
"user_id":   c.cfg.UserID,
"device_id": c.cfg.DeviceID,
}
helloBytes, _ := json.Marshal(hello)
msg1, hs, err := noisepkg.ClientStep1(clientCfg, helloBytes)
if err != nil {
return fmt.Errorf("handshake step1: %w", err)
}
if err := c.udp.Send(msg1); err != nil {
return fmt.Errorf("send handshake: %w", err)
}
buf := make([]byte, 65535)
c.udp.SetReadDeadline(time.Now().Add(10 * time.Second))
n, _, err := c.udp.Recv(buf)
if err != nil {
return fmt.Errorf("recv handshake reply: %w", err)
}
c.udp.SetReadDeadline(time.Time{})
noiseSess, payload, err := noisepkg.ClientStep2(hs, buf[:n])
if err != nil {
return fmt.Errorf("handshake step2: %w", err)
}
var reply struct {
Status     string `json:"status"`
AssignedIP string `json:"assigned_ip"`
}
if err := json.Unmarshal(payload, &reply); err != nil {
return fmt.Errorf("decode server reply: %w", err)
}
if reply.Status != "ok" || reply.AssignedIP == "" {
return fmt.Errorf("server rejected handshake: %s", string(payload))
}
ip, ipNet, err := net.ParseCIDR(reply.AssignedIP)
if err != nil {
ip = net.ParseIP(reply.AssignedIP)
if ip == nil {
return fmt.Errorf("invalid assigned ip: %q", reply.AssignedIP)
}
_, ipNet, _ = net.ParseCIDR(ip.String() + "/24")
}
ones, _ := ipNet.Mask.Size()
if err := c.tunDev.Configure(fmt.Sprintf("%s/%d", ip.String(), ones), net.IP(ipNet.Mask).String(), c.cfg.TUNMTU); err != nil {
return fmt.Errorf("configure tun: %w", err)
}
_ = c.tunDev.AddRoute("0.0.0.0/0")
c.session = &Session{
Noise:      noiseSess,
CreatedAt:  time.Now(),
LastSeen:   time.Now(),
AssignedIP: ip,
}
log.Printf("[client] tunnel up address=%s gateway=%s:%d", reply.AssignedIP, c.cfg.ServerHost, c.cfg.ServerPort)
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
if len(plaintext) == 0 {
continue
}
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
ciphertext, err := c.session.Noise.Encrypt(buf[:n])
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
