package tools

import (
	"fmt"
	"net"
	"os"
	"syscall"
	"time"

	"weavelab.xyz/ethr/lib"
)

func (t Tools) Dial(p lib.Protocol, dialAddr string, localIP net.IP, localPort uint16, ttl int, tos int) (net.Conn, error) {
	var lAddr net.Addr
	var network string
	var err error
	if p == lib.TCP {
		network = lib.TCPVersion(t.IPVersion)
		lAddr = &net.TCPAddr{
			IP:   localIP,
			Port: int(localPort),
		}
		//lAddr, err = net.ResolveTCPAddr(network, config.GetAddrString(localIP, localPort))
	} else if p == lib.UDP {
		network = lib.UDPVersion(t.IPVersion)
		lAddr = &net.UDPAddr{
			IP:   localIP,
			Port: int(localPort),
		}
		//lAddr, err = net.ResolveUDPAddr(network, config.GetAddrString(localIP, localPort))
	} else {
		return nil, fmt.Errorf("only TCP or UDP are allowed in dial: %w", os.ErrInvalid)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to resolve address: %w", err)
	}

	dialer := &net.Dialer{
		LocalAddr: lAddr,
		Timeout:   time.Second,
		Control: func(network, address string, rc syscall.RawConn) error {
			return rc.Control(func(fd uintptr) {
				_ = t.setTTL(fd, ttl, t.IPVersion)
				_ = t.setTOS(fd, tos, t.IPVersion)
			})
		},
	}
	conn, err := dialer.Dial(network, dialAddr)
	if err != nil {
		return nil, fmt.Errorf("error dialing remote: %w", err)
	}
	tcpConn, ok := conn.(*net.TCPConn)
	if ok {
		_ = tcpConn.SetLinger(0)
		return tcpConn, nil
	}
	udpConn, ok := conn.(*net.UDPConn)
	if ok {
		err = udpConn.SetWriteBuffer(4 * 1024 * 1024)
		if err != nil {
			return nil, fmt.Errorf("failed to set ReadBuffer on UDP socket: %w", err)
		}
		return udpConn, nil
	}

	return nil, fmt.Errorf("unknown connection type created")
}

func (t Tools) setTTL(fd uintptr, ttl int, ipVersion lib.IPVersion) error {
	if ttl == 0 {
		return nil
	}
	if ipVersion == lib.IPv4 {
		return t.setSockOptInt(fd, syscall.IPPROTO_IP, syscall.IP_TTL, ttl)
	}
	return t.setSockOptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_UNICAST_HOPS, ttl)
}

func (t Tools) setTOS(fd uintptr, tos int, ipVersion lib.IPVersion) error {
	if tos == 0 {
		return nil
	}
	if ipVersion == lib.IPv4 {
		return t.setSockOptInt(fd, syscall.IPPROTO_IP, syscall.IP_TOS, tos)
	}
	return t.setTClass(fd, tos)
}
