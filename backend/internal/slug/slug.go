// Package slug generates URL-safe identifiers from arbitrary strings.
package slug

import (
"crypto/rand"
"encoding/hex"
"strings"
"unicode"
)

const (
maxLen         = 60
randomSuffixLen = 6
)

// Generate converts a human-readable name into a URL-safe slug with a
// random suffix for uniqueness. The suffix prevents conflicts when two
// organizations or services share the same name.
//
// Example:
//
//Generate("ACME Corp") → "acme-corp-a3f9c2"
func Generate(name string) string {
base := normalize(name)
if base == "" {
base = "item"
}
if len(base) > maxLen {
base = strings.TrimRight(base[:maxLen], "-")
}
return base + "-" + randomSuffix()
}

// Normalize returns a deterministic slug without a random suffix.
// Useful for lookup keys that must be stable.
func Normalize(name string) string {
return normalize(name)
}

func normalize(s string) string {
s = strings.TrimSpace(s)
s = strings.ToLower(s)

var b strings.Builder
b.Grow(len(s))
prevDash := false

for _, r := range s {
switch {
case unicode.IsLetter(r) || unicode.IsDigit(r):
b.WriteRune(r)
prevDash = false
case r == '-' || r == '_' || r == ' ' || r == '.':
if !prevDash && b.Len() > 0 {
b.WriteByte('-')
prevDash = true
}
}
}

return strings.TrimRight(b.String(), "-")
}

func randomSuffix() string {
b := make([]byte, randomSuffixLen/2+1)
if _, err := rand.Read(b); err != nil {
return "000000"
}
return hex.EncodeToString(b)[:randomSuffixLen]
}
