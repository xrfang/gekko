package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"syscall"
	"unsafe"
)

const (
	c_IFNAMSIZ  = 16
	c_IFF_TUN   = 0x0001
	c_IFF_TAP   = 0x0002
	c_IFF_NO_PI = 0x1000
)

var ErrPermission error = fmt.Errorf("access denied")
var ErrDeviceName error = fmt.Errorf("invalid interface name")
var ErrDeviceBusy error = fmt.Errorf("device busy")

type ifReq struct {
	Name  [c_IFNAMSIZ]byte
	Flags uint16
	pad   [40 - c_IFNAMSIZ - 2]byte
}

func open(name string, flags uint16) (*os.File, error) {
	fd, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	if len(name) >= c_IFNAMSIZ {
		name = name[:c_IFNAMSIZ-1]
	}
	var req ifReq
	req.Flags = flags
	copy(req.Name[:], name)

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd.Fd(),
		uintptr(syscall.TUNSETIFF), uintptr(unsafe.Pointer(&req)))
	switch errno {
	case 0x01:
		return nil, ErrPermission
	case 0x10:
		return nil, ErrDeviceBusy
	case 0x16:
		if _, err := net.InterfaceByName(name); err == nil {
			return nil, ErrDeviceBusy
		}
		return nil, ErrDeviceName
	}

	i := bytes.IndexByte(req.Name[:], 0)
	return os.NewFile(fd.Fd(), string(req.Name[:i])), nil
}

// If name is empty, random device name will be used, like tun0, tun1, etc.
// fd.Name() is the real device name
func OpenTUN(name string) (fd *os.File, err error) {
	return open(name, c_IFF_TUN|c_IFF_NO_PI)
}

// If name is empty, random device name will be used, like tap0, tap1, etc.
// fd.Name() is the real device name
func OpenTAP(name string) (fd *os.File, err error) {
	return open(name, c_IFF_TAP|c_IFF_NO_PI)
}
