package transport

import (
	"fmt"
	"net"
	"time"
)

// UDP wraps a UDP socket for sending and receiving packets.
type UDP struct {
	conn *net.UDPConn
	peer *net.UDPAddr
}

// Listen binds a UDP socket to addr (server side).
func Listen(addr string) (*UDP, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	conn, err := net.ListenUDP("udp4", udpAddr)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	return &UDP{conn: conn}, nil
}

// Dial creates a connected UDP socket to addr (client side).
func Dial(addr string) (*UDP, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve: %w", err)
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	return &UDP{conn: conn, peer: udpAddr}, nil
}

// Send sends a datagram to the peer (client mode) or a specific addr (server mode).
func (u *UDP) Send(b []byte) error {
	if u.peer != nil {
		_, err := u.conn.WriteToUDP(b, u.peer)
		return err
	}
	_, err := u.conn.Write(b)
	return err
}

// SendTo sends a datagram to a specific UDP address (server mode).
func (u *UDP) SendTo(b []byte, addr *net.UDPAddr) error {
	_, err := u.conn.WriteToUDP(b, addr)
	return err
}

// Recv reads a datagram. Returns the bytes read and the sender's address.
func (u *UDP) Recv(buf []byte) (int, *net.UDPAddr, error) {
	n, addr, err := u.conn.ReadFromUDP(buf)
	return n, addr, err
}

// SetReadDeadline sets a read timeout.
func (u *UDP) SetReadDeadline(t time.Time) error {
	return u.conn.SetReadDeadline(t)
}

// Close closes the socket.
func (u *UDP) Close() error { return u.conn.Close() }

// LocalAddr returns the local socket address.
func (u *UDP) LocalAddr() net.Addr { return u.conn.LocalAddr() }
