// Simple implementation of an io.WriteCloser which counts how much data is
// written to it. It is best used in combination with either io.MultiWriter
// or io.TeeReader wrappers around an io.Writer or an io.Reader.
package counter

import (
	"github.com/dustin/go-humanize"
	"io"
	"log"
	"sync/atomic"
	"time"
)

var _ io.WriteCloser = &Counter{}

type Counter struct {
	prefix    string
	total     uint64
	count     uint64
	lastTotal [2]uint64
	lastCount [2]uint64
	started   time.Time
	close     chan struct{}
}

func NewCounter(prefix string) io.WriteCloser {
	if len(prefix) == 0 {
		prefix = "Write"
	}
	counter := &Counter{
		prefix:  prefix,
		started: time.Now(),
		close:   make(chan struct{}),
	}
	go counter.Reporter()
	return counter
}

func (wc *Counter) Write(p []byte) (n int, _ error) {
	n = len(p)
	_ = atomic.AddUint64(&wc.total, uint64(n))
	_ = atomic.AddUint64(&wc.count, 1)
	//log.Printf(
	//	"[DEBUG] count call: n: %s total: %s count: %s",
	//	humanize.IBytes(uint64(n)),
	//	humanize.IBytes(total),
	//	humanize.CommafWithDigits(float64(count), 0),
	//)
	return n, nil
}

func (wc *Counter) Close() error {
	select {
	case <-wc.close:
	default:
		close(wc.close)
	}
	return nil
}

func (wc *Counter) Reporter() {
	wc.close = make(chan struct{}, 1)
	ticker := time.NewTicker(time.Second / 2)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-wc.close:
			return

		case <-ticker.C:
			wc.report(i)
			if i == 0 {
				i = 1
			} else {
				i = 0
			}
		}
	}
}

func (wc *Counter) report(i int) {
	seconds := time.Since(wc.started).Seconds()
	total := atomic.LoadUint64(&wc.total)
	count := atomic.LoadUint64(&wc.count)

	lastTotal := min(wc.lastTotal[0], wc.lastTotal[1])
	lastCount := min(wc.lastCount[0], wc.lastCount[1])

	// Be silent if nothing has happened since last time
	if total == lastTotal && count == lastCount {
		return
	}

	wc.lastTotal[i] = total
	wc.lastCount[i] = count

	callsPerSecond := float64(count - lastCount)
	intervalBytesPerSecond := uint64(float64(total - lastTotal))
	averageBytesPerSecond := uint64(float64(total) / seconds)

	log.Printf(
		"[DEBUG] %16s %8s/s in %10s/s = total: %8s avg: %8s/s",
		wc.prefix,
		humanize.IBytes(intervalBytesPerSecond),
		humanize.CommafWithDigits(callsPerSecond, 0),
		humanize.IBytes(total),
		humanize.IBytes(averageBytesPerSecond),
	)
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}
