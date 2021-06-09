package client

import (
	"fmt"

	"weavelab.xyz/ethr/session/payloads"

	"weavelab.xyz/ethr/session"

	"weavelab.xyz/ethr/lib"
	"weavelab.xyz/ethr/ui"
)

func (u *UI) PrintBandwidth(test *session.Test, result *session.TestResult) {
	protocol := test.ID.Protocol
	if protocol != lib.TCP && protocol != lib.UDP {
		fmt.Printf("Unsupported protocol for bandwidth test: %s\n", protocol.String())
		return
	}

	if result.Body == nil {
		return
	}
	switch r := result.Body.(type) {
	case payloads.BandwidthPayload:
		if u.ShowConnectionStats {
			for _, conn := range r.ConnectionBandwidths {
				u.printBandwidthResult(protocol, conn.ConnectionID, conn.Bandwidth, conn.PacketsPerSecond)
			}
		}
		u.printBandwidthResult(protocol, "SUM", r.TotalBandwidth, r.TotalPacketsPerSecond)
		u.Logger.TestResult(lib.TestTypeBandwidth, true, protocol, test.RemoteIP, test.RemotePort, r)
	default:
		if r != nil {
			u.printUnknownResultType()
		}
	}
}

func (u *UI) PrintBandwidthHeader(p lib.Protocol) {
	if p == lib.UDP {
		// Printing packets only makes sense for UDP as it is a datagram protocol.
		// For TCP, TCP itself decides how to chunk the stream to send as packets.
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - -")
		fmt.Printf("%10s %12s %14s %10s %10s\n", "[  ID  ]", "Protocol", "Interval", "Bits/s", "Pkts/s")
	} else {
		fmt.Println("- - - - - - - - - - - - - - - - - - - - - - -")
		fmt.Printf("%10s %12s %14s %10s\n", "[  ID  ]", "Protocol", "Interval", "Bits/s")
	}
}

func (u *UI) printBandwidthResult(p lib.Protocol, id string, bw, pps uint64) {
	if p == lib.UDP {
		fmt.Printf("[%5s]     %-5s    %03d-%03d sec   %7s   %7s\n", id, p, u.lastPrintSeconds, u.currentPrintSeconds, ui.BytesToRate(bw), ui.PpsToString(pps))
	} else {
		fmt.Printf("[%5s]     %-5s    %03d-%03d sec   %7s\n", id, p, u.lastPrintSeconds, u.currentPrintSeconds, ui.BytesToRate(bw))
	}
}
