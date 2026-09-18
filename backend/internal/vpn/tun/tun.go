//go:build linux || windows

package tun

// Device is the common interface for TUN devices.
type Device interface {
	Name() string
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
	Configure(ip, mask string, mtu int) error
	AddRoute(cidr string) error
}

// Open creates a TUN device with the given name.
// The actual implementation is platform-specific.
func Open(name string) (Device, error) {
	return open(name)
}
