// +build darwin

package tools

import (
	"fmt"
	"net"
	"syscall"

	"weavelab.xyz/ethr/lib"
)

func (t Tools) setSockOptInt(fd uintptr, level, opt, val int) error {
	err := syscall.SetsockoptInt(int(fd), level, opt, val)
	if err != nil {
		return fmt.Errorf("failed to set socket option (%v) to value (%v): %w", opt, val, err)
	}
	return nil
}

func (t Tools) setTClass(fd uintptr, tos int) error {
	return t.setSockOptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_TCLASS, tos)
}

func (t Tools) IsAdmin() bool {
	return true
}

func (t Tools) IcmpNewConn(address string) (net.PacketConn, error) {
	dialedConn, err := net.Dial(lib.ICMPVersion(t.IPVersion), address)
	if err != nil {
		return nil, err
	}
	localAddr := dialedConn.LocalAddr()
	_ = dialedConn.Close()
	conn, err := net.ListenPacket(lib.ICMPVersion(t.IPVersion), localAddr.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}
