package engine

import (
"fmt"
"net"
"sync"
)

// IPPool hands out addresses from a fixed CIDR.
// Network and broadcast addresses are never assigned.
type IPPool struct {
mu       sync.Mutex
network  *net.IPNet
next     net.IP
assigned map[string]string
released []net.IP
}

func NewIPPool(cidr string, reserveFirst bool) (*IPPool, error) {
_, network, err := net.ParseCIDR(cidr)
if err != nil {
return nil, fmt.Errorf("parse cidr: %w", err)
}
ones, bits := network.Mask.Size()
if bits-ones < 2 {
return nil, fmt.Errorf("cidr too small: need at least /30")
}
p := &IPPool{
network:  network,
assigned: make(map[string]string),
}
p.next = incrementIP(network.IP)
if reserveFirst {
p.next = incrementIP(p.next)
}
return p, nil
}

func (p *IPPool) Allocate(sessionKey string) (net.IP, error) {
p.mu.Lock()
defer p.mu.Unlock()
if len(p.released) > 0 {
ip := p.released[len(p.released)-1]
p.released = p.released[:len(p.released)-1]
p.assigned[ip.String()] = sessionKey
return append(net.IP(nil), ip...), nil
}
for {
if !p.network.Contains(p.next) {
return nil, fmt.Errorf("address pool exhausted")
}
if isBroadcast(p.next, p.network) {
p.next = incrementIP(p.next)
continue
}
ipStr := p.next.String()
if _, taken := p.assigned[ipStr]; !taken {
p.assigned[ipStr] = sessionKey
out := append(net.IP(nil), p.next...)
p.next = incrementIP(p.next)
return out, nil
}
p.next = incrementIP(p.next)
}
}

func (p *IPPool) Release(ip net.IP) {
if ip == nil {
return
}
p.mu.Lock()
defer p.mu.Unlock()
ipStr := ip.String()
if _, ok := p.assigned[ipStr]; ok {
delete(p.assigned, ipStr)
p.released = append(p.released, append(net.IP(nil), ip...))
}
}

func incrementIP(ip net.IP) net.IP {
ip = append(net.IP(nil), ip...)
for i := len(ip) - 1; i >= 0; i-- {
ip[i]++
if ip[i] != 0 {
break
}
}
return ip
}

func isBroadcast(ip net.IP, network *net.IPNet) bool {
bcast := make(net.IP, len(network.IP))
for i := range network.IP {
bcast[i] = network.IP[i] | ^network.Mask[i]
}
return ip.Equal(bcast)
}
