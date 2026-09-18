package devices

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// KeyPair holds an X25519 keypair (used by Noise Protocol).
type KeyPair struct {
	PrivateKey []byte
	PublicKey  []byte
}

func GenerateKeyPair() (*KeyPair, error) {
	var privateKey [32]byte
	if _, err := rand.Read(privateKey[:]); err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}

	// Standard X25519 clamping
	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	publicKey, err := curve25519.X25519(privateKey[:], curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("derive public key: %w", err)
	}

	return &KeyPair{
		PrivateKey: privateKey[:],
		PublicKey:  publicKey,
	}, nil
}

func EncodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func DecodeKey(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
