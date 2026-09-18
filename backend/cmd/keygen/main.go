package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/curve25519"
)

func main() {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		fmt.Fprintf(os.Stderr, "rand: %v\n", err)
		os.Exit(1)
	}
	// X25519 clamping
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	pub, err := curve25519.X25519(priv[:], curve25519.Basepoint)
	if err != nil {
		fmt.Fprintf(os.Stderr, "derive: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server Keypair (X25519)")
	fmt.Println("-----------------------")
	fmt.Printf("Private: %s\n", base64.StdEncoding.EncodeToString(priv[:]))
	fmt.Printf("Public:  %s\n", base64.StdEncoding.EncodeToString(pub))
	fmt.Println()
	fmt.Println("Add to .env:")
	fmt.Printf("VPN_SERVER_PRIVATE_KEY=%s\n", base64.StdEncoding.EncodeToString(priv[:]))
	fmt.Printf("VPN_SERVER_PUBLIC_KEY=%s\n", base64.StdEncoding.EncodeToString(pub))
}
