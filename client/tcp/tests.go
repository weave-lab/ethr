package tcp

import (
	"weavelab.xyz/ethr/client/tools"
	"weavelab.xyz/ethr/lib"
)

type Tests struct {
	NetTools *tools.Tools
	Logger   lib.Logger // This is a hack figure out a better way
}
