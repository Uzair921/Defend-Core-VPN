package noise

import (
	"crypto/rand"
	"fmt"

	"github.com/flynn/noise"
)

// Session represents an established Noise session.
type Session struct {
	Send    *noise.CipherState
	Receive *noise.CipherState
}

// ClientHandshake performs the IK handshake as initiator.
// payload should contain the client's first message (can be empty).
// Returns the session and any payload received from the server.
func ClientHandshake(cfg noise.Config, payload []byte) (*Session, []byte, error) {
	hs, err := noise.NewHandshakeState(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("new handshake: %w", err)
	}

	// Message 1: client -> server
	msg1, _, _, err := hs.WriteMessage(nil, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("write msg1: %w", err)
	}

	// In a real transport, msg1 would be sent over UDP.
	// For this test, we return it so the caller can pass it to the server.
	// For now, we simulate a loopback: we don't actually send.
	// The caller (test code) handles the exchange.
	_ = msg1

	return nil, nil, fmt.Errorf("not implemented in loopback mode")
}

// ClientStep1 creates the first handshake message.
// Returns the message bytes and the handshake state (to continue later).
func ClientStep1(cfg noise.Config, payload []byte) ([]byte, *noise.HandshakeState, error) {
	hs, err := noise.NewHandshakeState(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("new handshake: %w", err)
	}

	msg, _, _, err := hs.WriteMessage(nil, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("write msg1: %w", err)
	}
	return msg, hs, nil
}

// ClientStep2 processes the server's reply and finishes the handshake.
func ClientStep2(hs *noise.HandshakeState, serverMsg []byte) (*Session, []byte, error) {
	payload, cs1, cs2, err := hs.ReadMessage(nil, serverMsg)
	if err != nil {
		return nil, nil, fmt.Errorf("read msg2: %w", err)
	}
	if cs1 == nil || cs2 == nil {
		return nil, nil, fmt.Errorf("handshake not complete")
	}
	return &Session{Send: cs1, Receive: cs2}, payload, nil
}

// ServerStep1 processes the client's first message.
func ServerStep1(cfg noise.Config, clientMsg []byte) (*noise.HandshakeState, []byte, error) {
	hs, err := noise.NewHandshakeState(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("new handshake: %w", err)
	}
	payload, _, _, err := hs.ReadMessage(nil, clientMsg)
	if err != nil {
		return nil, nil, fmt.Errorf("read msg1: %w", err)
	}
	return hs, payload, nil
}

// ServerStep2 creates the reply message and finishes the handshake.
func ServerStep2(hs *noise.HandshakeState, payload []byte) (*Session, []byte, error) {
	msg, cs1, cs2, err := hs.WriteMessage(nil, payload)
	if err != nil {
		return nil, nil, fmt.Errorf("write msg2: %w", err)
	}
	if cs1 == nil || cs2 == nil {
		return nil, nil, fmt.Errorf("handshake not complete")
	}
	// Server: cs1 is the server's SEND cipher (matches client's receive),
	// cs2 is the server's RECEIVE cipher (matches client's send).
	// Per Noise spec, initiator and responder have mirrored send/receive.
	_ = rand.Reader
	return &Session{Send: cs2, Receive: cs1}, msg, nil
}

// Encrypt encrypts plaintext using the session's send key.
func (s *Session) Encrypt(plaintext []byte) ([]byte, error) {
	if s.Send == nil {
		return nil, fmt.Errorf("send cipher not available")
	}
	return s.Send.Encrypt(nil, nil, plaintext)
}

// Decrypt decrypts ciphertext using the session's receive key.
func (s *Session) Decrypt(ciphertext []byte) ([]byte, error) {
	if s.Receive == nil {
		return nil, fmt.Errorf("receive cipher not available")
	}
	return s.Receive.Decrypt(nil, nil, ciphertext)
}
