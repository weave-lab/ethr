package session

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"weavelab.xyz/ethr/lib"
)

type Test struct {
	ID          lib.TestID
	IsActive    bool
	IsDormant   bool
	Session     *Session
	RemoteIP    net.IP
	RemotePort  uint16
	DialAddr    string
	ClientParam lib.ClientParams
	Results     chan TestResult
	Done        chan struct{}
	LastAccess  time.Time

	resultLock          sync.Mutex
	publishInterval     time.Duration
	intermediateResults []TestResult
	aggregator          ResultAggregator
	latestResult        *TestResult
}

type TestResult struct {
	Success bool
	Error   error
	Body    interface{}
}

type ResultAggregator func(uint64, []TestResult) TestResult

func NewTest(s *Session, protocol lib.Protocol, ttype lib.TestType, rIP net.IP, rPort uint16, params lib.ClientParams, aggregator ResultAggregator, publishInterval time.Duration) *Test {
	dialAddr := fmt.Sprintf("[%s]:%s", rIP.String(), strconv.Itoa(int(rPort)))
	if protocol == lib.ICMP {
		dialAddr = rIP.String()
	}
	return &Test{
		Session: s,
		ID: lib.TestID{
			Protocol: protocol,
			Type:     ttype,
		},
		RemoteIP:    rIP,
		RemotePort:  rPort,
		DialAddr:    dialAddr,
		ClientParam: params,
		Done:        make(chan struct{}),
		Results:     make(chan TestResult, 16),
		LastAccess:  time.Now(),
		IsDormant:   true,

		resultLock:          sync.Mutex{},
		publishInterval:     publishInterval,
		intermediateResults: make([]TestResult, 0, 100),
		aggregator:          aggregator,
		latestResult:        &TestResult{},
	}
}

func (t *Test) StartPublishing() {
	ticker := time.NewTicker(t.publishInterval) // most metrics are per second
	defer ticker.Stop()

	if t.aggregator != nil {
		t.republishAggregates(ticker)
	} else {
		t.republishAll(ticker)
	}
}

func (t *Test) republishAll(ticker *time.Ticker) {
	doRepublish := func() {
		t.resultLock.Lock()
		for _, r := range t.intermediateResults {
			select {
			case t.Results <- r:
			default:
			}
			t.latestResult = &r
		}
		if len(t.intermediateResults) > 0 {
			// TODO make sure old array is GC'ed
			t.intermediateResults = t.intermediateResults[:0]
			//t.intermediateResults = make([]TestResult, 0, cap(t.intermediateResults))
		}
		t.resultLock.Unlock()
	}

	for range ticker.C {
		select {
		case <-t.Done:
			doRepublish()
			close(t.Results)
			return
		default:
			doRepublish()
		}
	}
}

func (t *Test) republishAggregates(ticker *time.Ticker) {
	start := time.Now()

	doAggregate := func(start time.Time) bool {
		t.resultLock.Lock()
		if len(t.intermediateResults) == 0 {
			t.resultLock.Unlock()
			return false
		}

		ns := uint64(time.Since(start).Nanoseconds())
		if ns < 1 {
			ns = 1
		}
		r := t.aggregator(ns, t.intermediateResults)
		t.intermediateResults = make([]TestResult, 0, cap(t.intermediateResults))
		t.latestResult = &r
		t.resultLock.Unlock()

		select {
		case t.Results <- r:
		default:
		}
		return true
	}

	for range ticker.C {
		select {
		case <-t.Done:
			// cleanup any unpublished results
			_ = doAggregate(start)
			close(t.Results)
			return
		default:
			republished := doAggregate(start)
			if !republished {
				continue
			}
			start = time.Now()
		}
	}
}

func (t *Test) Terminate() {
	close(t.Done)
	t.IsActive = false

}

func (t *Test) AddIntermediateResult(r TestResult) {
	t.resultLock.Lock()
	defer t.resultLock.Unlock()
	t.intermediateResults = append(t.intermediateResults, r)
}

func (t *Test) LatestResult() *TestResult {
	t.resultLock.Lock()
	defer t.resultLock.Unlock()
	return t.latestResult
}

func (t *Test) AddDirectResult(r TestResult) {
	t.resultLock.Lock()
	defer t.resultLock.Unlock()
	t.latestResult = &r
	select {
	case t.Results <- r:
	default:
	}

}
