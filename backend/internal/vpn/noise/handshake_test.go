package noise

import (
	"bytes"
	"testing"
)

func TestHandshakeIK(t *testing.T) {
	// Server keypair (long-lived, would be in .env)
	serverKP, err := GenerateStaticKeypair()
	if err != nil {
		t.Fatalf("server keypair: %v", err)
	}

	// Server config
	serverCfg, err := ServerConfig(serverKP)
	if err != nil {
		t.Fatalf("server config: %v", err)
	}

	// Client config (knows server's public key)
	clientCfg, err := ClientConfig(serverKP.Public)
	if err != nil {
		t.Fatalf("client config: %v", err)
	}

	// ===== Step 1: Client sends msg1 =====
	clientMsg1, clientHS, err := ClientStep1(clientCfg, []byte("hello from client"))
	if err != nil {
		t.Fatalf("client step1: %v", err)
	}
	t.Logf("client msg1: %d bytes", len(clientMsg1))

	// ===== Step 2: Server processes msg1, sends msg2 =====
	serverHS, serverPayload, err := ServerStep1(serverCfg, clientMsg1)
	if err != nil {
		t.Fatalf("server step1: %v", err)
	}
	if !bytes.Equal(serverPayload, []byte("hello from client")) {
		t.Fatalf("server payload mismatch: %q", serverPayload)
	}
	t.Logf("server received: %q", serverPayload)

	serverSession, serverMsg2, err := ServerStep2(serverHS, []byte("hello from server"))
	if err != nil {
		t.Fatalf("server step2: %v", err)
	}
	t.Logf("server msg2: %d bytes", len(serverMsg2))

	// ===== Step 3: Client processes msg2 =====
	clientSession, clientPayload, err := ClientStep2(clientHS, serverMsg2)
	if err != nil {
		t.Fatalf("client step2: %v", err)
	}
	if !bytes.Equal(clientPayload, []byte("hello from server")) {
		t.Fatalf("client payload mismatch: %q", clientPayload)
	}
	t.Logf("client received: %q", clientPayload)

	// ===== Step 4: Encrypted transport (client -> server) =====
	plaintext := []byte("secret data from client")
	ciphertext, err := clientSession.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("client encrypt: %v", err)
	}
	t.Logf("ciphertext: %d bytes", len(ciphertext))

	decrypted, err := serverSession.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("server decrypt: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("server decrypt mismatch: %q", decrypted)
	}
	t.Logf("server decrypted: %q", decrypted)

	// ===== Step 5: Encrypted transport (server -> client) =====
	plaintext2 := []byte("secret data from server")
	ciphertext2, err := serverSession.Encrypt(plaintext2)
	if err != nil {
		t.Fatalf("server encrypt: %v", err)
	}
	decrypted2, err := clientSession.Decrypt(ciphertext2)
	if err != nil {
		t.Fatalf("client decrypt: %v", err)
	}
	if !bytes.Equal(decrypted2, plaintext2) {
		t.Fatalf("client decrypt mismatch: %q", decrypted2)
	}
	t.Logf("client decrypted: %q", decrypted2)

	t.Log("✅ Handshake + bidirectional encryption OK")
}
