package noise

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/flynn/noise"
)

// Noise pattern: IK
//   I = initiator (client) knows responder's (server's) static public key
//   K = static key known at start
//
// Cipher suite: X25519 + ChaCha20-Poly1305 + BLAKE2s
// Exactly what WireGuard uses.

var (
	Pattern     = noise.HandshakeIK
	CipherSuite = noise.CipherChaChaPoly
	Hash        = noise.HashBLAKE2s
	DH          = noise.DH25519
)

// StaticKeypair holds a Noise static keypair.
type StaticKeypair struct {
	Private []byte
	Public  []byte
}

// DecodeStaticKeypair from base64 strings.
func DecodeStaticKeypair(privB64, pubB64 string) (*StaticKeypair, error) {
	priv, err := base64.StdEncoding.DecodeString(privB64)
	if err != nil {
		return nil, fmt.Errorf("decode private: %w", err)
	}
	pub, err := base64.StdEncoding.DecodeString(pubB64)
	if err != nil {
		return nil, fmt.Errorf("decode public: %w", err)
	}
	if len(priv) != 32 || len(pub) != 32 {
		return nil, fmt.Errorf("invalid key length: priv=%d pub=%d", len(priv), len(pub))
	}
	return &StaticKeypair{Private: priv, Public: pub}, nil
}

// GenerateStaticKeypair for testing / first-time setup.
func GenerateStaticKeypair() (*StaticKeypair, error) {
	kp, err := noise.DH25519.GenerateKeypair(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &StaticKeypair{
		Private: kp.Private,
		Public:  kp.Public,
	}, nil
}

// Encode returns base64 representations.
func (k *StaticKeypair) Encode() (privB64, pubB64 string) {
	return base64.StdEncoding.EncodeToString(k.Private),
		base64.StdEncoding.EncodeToString(k.Public)
}

// NewCipherSuite returns the Noise cipher suite we use.
func NewCipherSuite() noise.CipherSuite {
	return noise.NewCipherSuite(DH, CipherSuite, Hash)
}

// ServerConfig builds the responder config for the server.
// The server's static keypair is long-lived.
func ServerConfig(static *StaticKeypair) (noise.Config, error) {
	cs := NewCipherSuite()

	return noise.Config{
		CipherSuite:   cs,
		Random:        rand.Reader,
		Pattern:       Pattern,
		Initiator:     false, // server is responder
		StaticKeypair: noise.DHKey{Private: static.Private, Public: static.Public},
	}, nil
}

// ClientConfig builds the initiator config for the client.
// ServerPub is the responder's static public key.
func ClientConfig(serverPub []byte) (noise.Config, error) {
	cs := NewCipherSuite()

	clientStatic, err := cs.GenerateKeypair(rand.Reader)
	if err != nil {
		return noise.Config{}, fmt.Errorf("client static keypair: %w", err)
	}

	return noise.Config{
		CipherSuite:   cs,
		Random:        rand.Reader,
		Pattern:       Pattern,
		Initiator:     true,
		StaticKeypair: noise.DHKey{Private: clientStatic.Private, Public: clientStatic.Public},
		PeerStatic:    serverPub,
	}, nil
}
