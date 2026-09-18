//go:build linux

package tun

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"
)

const (
	IFF_TUN   = 0x0001
	IFF_NO_PI = 0x1000
	TUNSETIFF = 0x400454ca
)

type linuxTUN struct {
	name string
	file *os.File
}

type ifreq struct {
	Name  [16]byte
	Flags uint16
	_     [22]byte
}

func open(name string) (Device, error) {
	if unsafe.Sizeof(ifreq{}) != 40 {
		return nil, fmt.Errorf("ifreq size mismatch: got %d, want 40", unsafe.Sizeof(ifreq{}))
	}

	fd, err := syscall.Open("/dev/net/tun", syscall.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open /dev/net/tun: %w", err)
	}

	var req ifreq
	copy(req.Name[:15], name)
	req.Flags = IFF_TUN | IFF_NO_PI

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(fd),
		uintptr(TUNSETIFF),
		uintptr(unsafe.Pointer(&req)),
	)
	if errno != 0 {
		syscall.Close(fd)
		return nil, fmt.Errorf("ioctl TUNSETIFF: %v", errno)
	}

	actualName := strings.TrimRight(string(req.Name[:]), "\x00")
	return &linuxTUN{
		name: actualName,
		file: os.NewFile(uintptr(fd), "/dev/net/tun"),
	}, nil
}

func (t *linuxTUN) Name() string { return t.name }

func (t *linuxTUN) Read(p []byte) (int, error)  { return t.file.Read(p) }
func (t *linuxTUN) Write(p []byte) (int, error) { return t.file.Write(p) }

func (t *linuxTUN) Close() error { return t.file.Close() }

func (t *linuxTUN) Configure(ip, mask string, mtu int) error {
	if err := run("ip", "addr", "add", ip, "dev", t.name); err != nil {
		return fmt.Errorf("ip addr add: %w", err)
	}
	if err := run("ip", "link", "set", "dev", t.name, "mtu", fmt.Sprintf("%d", mtu)); err != nil {
		return fmt.Errorf("ip link set mtu: %w", err)
	}
	if err := run("ip", "link", "set", "dev", t.name, "up"); err != nil {
		return fmt.Errorf("ip link set up: %w", err)
	}
	return nil
}

func (t *linuxTUN) AddRoute(cidr string) error {
	return run("ip", "route", "add", cidr, "dev", t.name)
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v: %w (output: %s)", name, args, err, string(out))
	}
	return nil
}
