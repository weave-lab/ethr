package tools

import (
	"net"

	"weavelab.xyz/ethr/lib"
)

type Tools struct {
	IPVersion lib.IPVersion
	Logger    lib.Logger

	IsExternal bool
	RemoteIP   net.IP
	RemotePort uint16

	LocalPort uint16
	LocalIP   net.IP
}

func NewTools(isExternal bool, rIP net.IP, rPort uint16, localPort uint16, localIP net.IP, logger lib.Logger) (*Tools, error) {
	var ipVersion lib.IPVersion
	if rIP != nil {
		if rIP.To4() != nil {
			ipVersion = lib.IPv4
		} else {
			ipVersion = lib.IPv6
		}
	}
	//else {
	//	return nil, fmt.Errorf("failed to parse server IP from (%s)", rIP)
	//}

	return &Tools{
		IPVersion:  ipVersion,
		IsExternal: isExternal,
		RemoteIP:   rIP,
		RemotePort: rPort,
		LocalPort:  localPort,
		LocalIP:    localIP,
		Logger:     logger,
	}, nil
}
