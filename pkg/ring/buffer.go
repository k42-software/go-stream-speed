// Adapted from https://github.com/egonelbre/exp/tree/master/ring
// UNLICENSE https://github.com/egonelbre/exp/blob/master/LICENSE

// Simple Ring Buffer using a Bounded MPMC queue.
package ring

import (
	"sync"
	"sync/atomic"
)

// Implementation is based on
//   http://www.1024cores.net/home/lock-free-algorithms/queues/bounded-mpmc-queue

const (
	Size     = 256 << 10 // original: 128 << 10
	Mask     = Size - 1
	CellSize = 4 << 10 // 4kb

	// setting this to a lower value has a minor performance impact
	// this is very workload dependant - your mileage may vary
	busyWaitIter = 128
)

type Cell struct {
	sequence int64
	Size     int64
	Data     [CellSize]byte
	wake     func()
}

type Ring struct {
	writeat  int64
	readfrom int64
	buffer   [Size]Cell
	signal   *sync.Cond
}

// Ring Buffer
func New() *Ring {
	ring := &Ring{
		signal: sync.NewCond(&sync.Mutex{}),
	}
	for i := range ring.buffer {
		ring.buffer[i].sequence = int64(i)
		ring.buffer[i].wake = ring.wake
	}
	return ring
}

func (ring *Ring) wake() {
	ring.signal.Broadcast()
}

func (ring *Ring) wait() {
	ring.signal.L.Lock()
	ring.signal.Wait()
	ring.signal.L.Unlock()
}

func (ring *Ring) Allocate() *Cell {
	busywait := busyWaitIter
	for {
		pos := atomic.LoadInt64(&ring.writeat)
		cell := &ring.buffer[pos&Mask]
		seq := atomic.LoadInt64(&cell.sequence)
		if seq-pos == 0 {
			if atomic.CompareAndSwapInt64(&ring.writeat, pos, pos+1) {
				return cell
			}
		}

		if busywait < 0 {
			ring.wait() // This is much faster than using runtime.Gosched()
		} else {
			busywait--
		}
	}
}

func (cell *Cell) Enqueue() {
	atomic.AddInt64(&cell.sequence, 1)
	cell.wake()
}

func (ring *Ring) Dequeue() (cell *Cell) {
	busywait := busyWaitIter
	for {
		pos := atomic.LoadInt64(&ring.readfrom)
		cell = &ring.buffer[pos&Mask]
		seq := atomic.LoadInt64(&cell.sequence)
		dif := seq - (pos + 1)
		if dif == 0 {
			if atomic.CompareAndSwapInt64(&ring.readfrom, pos, pos+1) {
				return cell
			}
		}

		if busywait < 0 {
			ring.wait() // This is much faster than using runtime.Gosched()
		} else {
			busywait--
		}
	}
}

func (cell *Cell) Release() {
	atomic.AddInt64(&cell.sequence, Size-1)
	cell.wake()
}
