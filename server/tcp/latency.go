package tcp

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"weavelab.xyz/ethr/lib"
	"weavelab.xyz/ethr/session"
	"weavelab.xyz/ethr/session/payloads"
)

func (h Handler) TestLatency(ctx context.Context, test *session.Test, clientParam lib.ClientParams, conn net.Conn) error {
	bytes := make([]byte, clientParam.BufferSize)
	rttCount := clientParam.RttCount
	latencyNumbers := make([]time.Duration, rttCount)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		_, err := io.ReadFull(conn, bytes)
		if err != nil {
			return fmt.Errorf("error receiving data for latency tests: %w", err)
		}
		for i := uint32(0); i < rttCount; i++ {
			s1 := time.Now()
			_, err = conn.Write(bytes)
			if err != nil {
				return fmt.Errorf("error sending data for latency test: %w", err)

			}
			_, err = io.ReadFull(conn, bytes)
			if err != nil {
				return fmt.Errorf("error receiving data for latency test: %w", err)

			}
			e2 := time.Since(s1)
			latencyNumbers[i] = e2
		}

		test.AddIntermediateResult(session.TestResult{
			Success: true,
			Error:   nil,
			Body:    payloads.RawLatencies{Latencies: latencyNumbers},
		})
	}
}
