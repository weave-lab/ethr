package server

import (
	"net"

	"weavelab.xyz/ethr/lib"
)

type Config struct {
	IPVersion lib.IPVersion
	LocalIP   net.IP
	LocalPort uint16
}
