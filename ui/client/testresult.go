package client

import (
	"context"
	"fmt"
	"time"

	"weavelab.xyz/ethr/lib"

	"weavelab.xyz/ethr/session"
)

func (u *UI) PrintTestResults(ctx context.Context, test *session.Test) {
	started := time.Now()
	exiting := false
	displayedHeader := false
	var previousResult, latestResult *session.TestResult

	tickInterval := 250 * time.Millisecond
	if test.ID.Type == lib.TestTypeBandwidth || test.ID.Type == lib.TestTypePacketsPerSecond || test.ID.Type == lib.TestTypeConnectionsPerSecond {
		tickInterval = time.Second
	}
	paintTicker := time.NewTicker(tickInterval)

	for {
		// TODO probably want this a little more accurate than rounded seconds, some things it makes sense to print more than once a second
		u.currentPrintSeconds = uint64(time.Since(started).Seconds())
		switch test.ID.Type {
		case lib.TestTypePing:
			if latestResult != previousResult && latestResult != nil {
				u.PrintPing(test, latestResult)
			}
		case lib.TestTypePacketsPerSecond:
			if !displayedHeader {
				u.PrintPacketsPerSecondHeader()
				displayedHeader = true
			}
			if latestResult != previousResult && latestResult != nil {
				u.PrintPacketsPerSecond(test, latestResult)
			}
		case lib.TestTypeBandwidth:
			if !displayedHeader {
				u.PrintBandwidthHeader(test.ID.Protocol)
				displayedHeader = true
			}
			if latestResult != previousResult && latestResult != nil {
				u.PrintBandwidth(test, latestResult)
			}
		case lib.TestTypeLatency:
			if !displayedHeader {
				u.PrintLatencyHeader()
				displayedHeader = true
			}
			if latestResult != previousResult && latestResult != nil {
				u.PrintLatency(test, latestResult)
			}
		case lib.TestTypeConnectionsPerSecond:
			if !displayedHeader {
				u.PrintConnectionsHeader()
				displayedHeader = true
			}
			if latestResult != previousResult && latestResult != nil {
				u.PrintConnectionsPerSecond(test, latestResult)
			}
		case lib.TestTypeTraceRoute:
			fallthrough
		case lib.TestTypeMyTraceRoute:
			if !displayedHeader {
				u.PrintTracerouteHeader(test.RemoteIP)
				displayedHeader = true
			}
			// if we are exiting drain the results to make sure everything gets printed
			if exiting {
				for r := range test.Results {
					u.PrintTraceroute(test, &r)
				}
				return
			}

			select {
			case r := <-test.Results:
				u.PrintTraceroute(test, &r)
			default:
			}
		default:
			u.printUnknownResultType()
		}
		// TODO probably want this a little more accurate, some things it makes sense to print more than once a second
		u.lastPrintSeconds = uint64(time.Since(started).Seconds())

		select {
		case <-paintTicker.C:
			// TODO convert each test type to read from results chan
			previousResult = latestResult
			latestResult = test.LatestResult()
			continue
		case <-test.Done:
			// Ensure one last paint
			if exiting {
				return
			}
			exiting = true
		case <-ctx.Done():
			return
		}
	}
}

func (u *UI) printUnknownResultType() {
	fmt.Printf("Unknown test result...\n")
}
