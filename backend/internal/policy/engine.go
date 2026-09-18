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
	return &Engine{rules: rules}
}

// CheckDestination checks if a destination IP/domain is allowed.
// Returns (allowed, matchedRule).
func (e *Engine) CheckDestination(ip string, domain string) (bool, *AccessPolicy) {
	for i := range e.rules {
		rule := &e.rules[i]
		if e.matches(rule, ip, domain) {
			return rule.Action == PolicyActionAllow, rule
		}
	}
	// Zero-trust default: DENY
	return false, nil
}

func (e *Engine) matches(rule *AccessPolicy, ip string, domain string) bool {
	switch rule.Type {
	case PolicyTypeCIDR:
		_, cidr, err := net.ParseCIDR(rule.Value)
		if err != nil {
			return false
		}
		destIP := net.ParseIP(ip)
		return destIP != nil && cidr.Contains(destIP)

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
		// Port-based handled elsewhere; matches only if port matches
		return false
	}
	return false
}

// GenerateFirewallRules generates iptables rules for a client IP.
func (e *Engine) GenerateFirewallRules(clientIP string) []string {
	rules := []string{
		fmt.Sprintf("# DefendCore policy for %s", clientIP),
		// Allow established
		fmt.Sprintf("iptables -A FORWARD -s %s -m state --state ESTABLISHED,RELATED -j ACCEPT", clientIP),
	}

	for _, r := range e.rules {
		if r.Type != PolicyTypeCIDR && r.Type != PolicyTypeIP {
			continue
		}
		action := "ACCEPT"
		if r.Action == PolicyActionDeny {
			action = "DROP"
		}
		target := r.Value
		rules = append(rules, fmt.Sprintf("iptables -A FORWARD -s %s -d %s -j %s", clientIP, target, action))
	}

	// Default deny
	rules = append(rules, fmt.Sprintf("iptables -A FORWARD -s %s -j DROP", clientIP))
	return rules
}

// GenerateCleanupRules removes policy rules for a client IP.
func (e *Engine) GenerateCleanupRules(clientIP string) []string {
	return []string{
		fmt.Sprintf("# Cleanup DefendCore policy for %s", clientIP),
		fmt.Sprintf("iptables -D FORWARD -s %s -j DROP 2>/dev/null || true", clientIP),
		// Add more -D rules as needed
	}
}
