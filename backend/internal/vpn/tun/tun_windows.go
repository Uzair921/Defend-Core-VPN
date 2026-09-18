//go:build windows && cgo

package tun

/*
#include <windows.h>
*/
import "C"

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
	"runtime"
)

var (
	wintunDLL                syscall.Handle
	procCreateAdapter        uintptr
	procOpenAdapter          uintptr
	procCloseAdapter         uintptr
	procStartSession         uintptr
	procEndSession           uintptr
	procReceivePacket        uintptr
	procReleaseReceivePacket uintptr
	procAllocateSendPacket   uintptr
	procSendPacket           uintptr
	procGetReadWaitEvent     uintptr
)

func loadWintun() error {
	if wintunDLL != 0 {
		return nil
	}
	dll, err := syscall.LoadLibrary("wintun.dll")
	if err != nil {
		return fmt.Errorf("LoadLibrary wintun.dll: %w", err)
	}
	wintunDLL = dll

	procs := map[string]*uintptr{
		"WintunCreateAdapter":        &procCreateAdapter,
		"WintunOpenAdapter":          &procOpenAdapter,
		"WintunCloseAdapter":         &procCloseAdapter,
		"WintunStartSession":         &procStartSession,
		"WintunEndSession":           &procEndSession,
		"WintunReceivePacket":        &procReceivePacket,
		"WintunReleaseReceivePacket": &procReleaseReceivePacket,
		"WintunAllocateSendPacket":   &procAllocateSendPacket,
		"WintunSendPacket":           &procSendPacket,
		"WintunGetReadWaitEvent":     &procGetReadWaitEvent,
	}
	for name, ptr := range procs {
		addr, err := syscall.GetProcAddress(dll, name)
		if err != nil {
			return fmt.Errorf("GetProcAddress %s: %w", name, err)
		}
		*ptr = addr
	}
	log.Printf("[wintun] all function pointers loaded")
	return nil
}

type windowsTUN struct {
	name         string
	adapter      uintptr
	session      uintptr
	readWait     uintptr
	interfaceIdx int
}

func open(name string) (Device, error) {
	log.Printf("[wintun] open(%q)", name)
	if err := loadWintun(); err != nil {
		return nil, err
	}

	namePtr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return nil, err
	}
	typePtr, err := syscall.UTF16PtrFromString("DefendCore")
	if err != nil {
		return nil, err
	}

	// Try open existing
	r1, _, _ := syscall.SyscallN(procOpenAdapter, uintptr(unsafe.Pointer(namePtr)))
	log.Printf("[wintun] WintunOpenAdapter(%q) = 0x%x", name, r1)

	var adapter uintptr
	if r1 != 0 {
		adapter = r1
		log.Printf("[wintun] opened existing adapter")
	} else {
		r1, _, errno := syscall.SyscallN(
			procCreateAdapter,
			uintptr(unsafe.Pointer(namePtr)),
			uintptr(unsafe.Pointer(typePtr)),
			0,
		)
		log.Printf("[wintun] WintunCreateAdapter(%q, DefendCore) = 0x%x, errno=%d", name, r1, errno)
		if r1 == 0 {
			return nil, fmt.Errorf("WintunCreateAdapter failed: errno=%d", errno)
		}
		adapter = r1
	}

	// Set *IfType = 1 on Wintun driver registry (point-to-point mode)
	// This makes Windows skip ARP on the Wintun adapter.
	if runtime.GOOS == "windows" {
		psCmd := `$wintunKeys = Get-ChildItem "HKLM:\SYSTEM\CurrentControlSet\Control\Class\{4d36e972-e325-11ce-bfc1-08002be10318}" -ErrorAction SilentlyContinue | Where-Object { (Get-ItemProperty $_.PSPath -ErrorAction SilentlyContinue).DriverDesc -eq "Wintun Userspace Tunnel" }; foreach ($key in $wintunKeys) { Set-ItemProperty -Path $key.PSPath -Name "*IfType" -Value 1 -Type DWORD -ErrorAction SilentlyContinue }`
		_ = exec.Command("powershell", "-Command", psCmd).Run()
		time.Sleep(1 * time.Second)
	}

	session, _, errno := syscall.SyscallN(procStartSession, adapter, 0x400000)
	log.Printf("[wintun] WintunStartSession = 0x%x, errno=%d", session, errno)
	if session == 0 {
		syscall.SyscallN(procCloseAdapter, adapter)
		return nil, fmt.Errorf("WintunStartSession failed: errno=%d", errno)
	}

	readWait, _, _ := syscall.SyscallN(procGetReadWaitEvent, session)
	log.Printf("[wintun] WintunGetReadWaitEvent = 0x%x", readWait)

	return &windowsTUN{
		name:     name,
		adapter:  adapter,
		session:  session,
		readWait: readWait,
	}, nil
}

func (t *windowsTUN) Name() string { return t.name }

func (t *windowsTUN) Read(p []byte) (int, error) {
	var packetSize uint32
	for {
		syscall.WaitForSingleObject(syscall.Handle(t.readWait), syscall.INFINITE)
		packet, _, _ := syscall.SyscallN(
			procReceivePacket,
			t.session,
			uintptr(unsafe.Pointer(&packetSize)),
		)
		if packet != 0 {
			n := int(packetSize)
			if n > len(p) {
				n = len(p)
			}
			copy(p[:n], unsafe.Slice((*byte)(unsafe.Pointer(packet)), n))
			syscall.SyscallN(procReleaseReceivePacket, t.session, packet)
			return n, nil
		}
	}
}

func (t *windowsTUN) Write(p []byte) (int, error) {
	packet, _, errno := syscall.SyscallN(
		procAllocateSendPacket,
		t.session,
		uintptr(len(p)),
	)
	if packet == 0 {
		return 0, fmt.Errorf("WintunAllocateSendPacket failed: %v", errno)
	}
	copy(unsafe.Slice((*byte)(unsafe.Pointer(packet)), len(p)), p)
	syscall.SyscallN(procSendPacket, t.session, packet)
	return len(p), nil
}

func (t *windowsTUN) Close() error {
	if t.session != 0 {
		syscall.SyscallN(procEndSession, t.session)
	}
	if t.adapter != 0 {
		syscall.SyscallN(procCloseAdapter, t.adapter)
	}
	return nil
}

// findInterfaceName finds the actual Windows interface name for our Wintun adapter.
// Wintun creates the adapter with a name, but Windows may append a suffix or use a different name.
func (t *windowsTUN) findInterfaceName() (string, error) {
	out, err := exec.Command("netsh", "interface", "show", "interface").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("netsh show interface: %w", err)
	}

	// Look for our adapter name in the output
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, t.name) {
			// Format: "Enabled  Connected  Dedicated  InterfaceName"
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				// Last field is the interface name (may have spaces)
				idx := strings.Index(line, fields[3])
				if idx >= 0 {
					return strings.TrimSpace(line[idx:]), nil
				}
			}
		}
	}
	return "", fmt.Errorf("interface %q not found in netsh output", t.name)
}

func (t *windowsTUN) Configure(ip, mask string, mtu int) error {
	log.Printf("[wintun] Configure: name=%q ip=%s mask=%s mtu=%d", t.name, ip, mask, mtu)

	// Wait a moment for Windows to register the adapter
	// Wintun adapter creation is async — Windows needs time to register
	for i := 0; i < 10; i++ {
		out, err := exec.Command("netsh", "interface", "show", "interface").CombinedOutput()
		if err == nil && strings.Contains(string(out), t.name) {
			log.Printf("[wintun] adapter found in netsh output after %d attempts", i+1)
			break
		}
		if i == 9 {
			log.Printf("[wintun] WARNING: adapter %q not found after 10 attempts", t.name)
			log.Printf("[wintun] netsh output:\n%s", string(out))
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Set IP
	if err := runNetsh("interface", "ip", "set", "address", "name="+t.name, "static", ip, "255.255.255.255"); err != nil {
		log.Printf("[wintun] set IP failed: %v", err)
		// Try with dhcp first, then static
		runNetsh("interface", "ip", "set", "address", "name="+t.name, "dhcp")
		time.Sleep(500 * time.Millisecond)
		if err2 := runNetsh("interface", "ip", "set", "address", "name="+t.name, "static", ip, "255.255.255.255"); err2 != nil {
			return fmt.Errorf("set IP (retry): %w", err2)
		}
	}

	// Set MTU
	if err := runNetsh("interface", "ipv4", "set", "subinterface", t.name,
		fmt.Sprintf("mtu=%d", mtu), "store=persistent"); err != nil {
		log.Printf("[wintun] set MTU failed: %v", err)
		// non-fatal
	}
	// Add static ARP entry for server (Wintun doesn't do ARP)
	_ = runNetsh("interface", "ipv4", "add", "neighbors",
		t.name, "10.8.0.1", "02-00-00-00-00-01")

	// Set interface metric
	_ = runNetsh("interface", "ipv4", "set", "interface", t.name, "metric=1")

	return nil

}

func (t *windowsTUN) AddRoute(cidr string) error {
	log.Printf("[wintun] AddRoute: %s via %q", cidr, t.name)
	return runNetsh("interface", "ipv4", "add", "route", cidr, t.name)
}

func runNetsh(args ...string) error {
	out, err := exec.Command("netsh", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("netsh %v: %w (output: %s)", args, err, string(out))
	}
	return nil
}
