package policy

import (
"fmt"
"net"
"strings"
)

type Engine struct {
rules []AccessPolicy
}

func NewEngine(rules []AccessPolicy) *Engine {
sorted := make([]AccessPolicy, len(rules))
copy(sorted, rules)
for i := 0; i < len(sorted); i++ {
for j := i + 1; j < len(sorted); j++ {
if sorted[j].Priority > sorted[i].Priority {
sorted[i], sorted[j] = sorted[j], sorted[i]
}
}
}
return &Engine{rules: sorted}
}

func (e *Engine) CheckDestination(ip, domain string) (allowed bool, matched *AccessPolicy) {
for i := range e.rules {
rule := &e.rules[i]
if !e.matches(rule, ip, domain) {
continue
}
return rule.Action == PolicyActionAllow, rule
}
return false, nil
}

func (e *Engine) matches(rule *AccessPolicy, ip, domain string) bool {
switch rule.Type {
case PolicyTypeCIDR:
_, cidr, err := net.ParseCIDR(rule.Value)
if err != nil {
return false
}
dst := net.ParseIP(ip)
return dst != nil && cidr.Contains(dst)
case PolicyTypeIP:
return rule.Value == ip
case PolicyTypeDomain:
if domain == "" {
return false
}
d := strings.ToLower(domain)
v := strings.ToLower(rule.Value)
return d == v || strings.HasSuffix(d, "."+v)
case PolicyTypeURL:
if domain == "" {
return false
}
return strings.Contains(strings.ToLower(rule.Value), strings.ToLower(domain))
case PolicyTypePort:
return false
}
return false
}

func (e *Engine) GenerateFirewallRules(clientIP string) ([]string, error) {
if net.ParseIP(clientIP) == nil {
return nil, fmt.Errorf("invalid client IP: %q", clientIP)
}
out := []string{
fmt.Sprintf("# DefendCore policy for %s", clientIP),
fmt.Sprintf("iptables -A FORWARD -s %s -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT", clientIP),
}
for _, r := range e.rules {
if r.Type != PolicyTypeCIDR && r.Type != PolicyTypeIP {
continue
}
if !validNetworkToken(r.Value) {
continue
}
action := "ACCEPT"
if r.Action == PolicyActionDeny {
action = "DROP"
}
out = append(out, fmt.Sprintf("iptables -A FORWARD -s %s -d %s -j %s", clientIP, r.Value, action))
}
out = append(out, fmt.Sprintf("iptables -A FORWARD -s %s -j DROP", clientIP))
return out, nil
}

func validNetworkToken(v string) bool {
if ip := net.ParseIP(v); ip != nil {
return true
}
_, _, err := net.ParseCIDR(v)
return err == nil
}
