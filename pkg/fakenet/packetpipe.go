package fakenet

import (
	"fmt"
	"github.com/k42-software/go-stream-speed/pkg/ring"
	"io"
	"log"
	"sync"
	"sync/atomic"
)

var _ io.ReadWriteCloser = &packetPipe{}

// Fake/Virtual Point To Point Simulated Network Link
type NetworkPipe interface {
	io.ReadWriteCloser
	MTU() int64
}

// Asynchronous packet based NetworkPipe backed by a ring buffer
type packetPipe struct {
	ringBuffer *ring.Ring
	mtu        int64

	// close management
	writeClosed int32
	readClosed  int32
	closeOnce   sync.Once
}

// Asynchronous packet based NetworkPipe backed by a ring buffer
func NewPacketPipe(mtu int64) (pipe NetworkPipe, err error) {
	if mtu == 0 {
		mtu = StandardMTU
	}
	if mtu > MaximumMTU {
		err = fmt.Errorf("mtu too large: %d", mtu)
		mtu = MaximumMTU
	}
	if mtu <= MinimumMTU {
		err = fmt.Errorf("mtu too small: %d", mtu)
		mtu = MinimumMTU
	}
	pipe = &packetPipe{
		ringBuffer: ring.New(),
		mtu:        mtu,
	}
	return pipe, err
}

func (r *packetPipe) MTU() int64 {
	return r.mtu
}

func (r *packetPipe) Read(p []byte) (n int, err error) {

	if atomic.LoadInt32(&r.readClosed) > 0 {
		return 0, io.EOF
	}
	//log.Printf("[DEBUG] fakenet: innerPipe.Read p:%d", len(p))

	cell := r.ringBuffer.Dequeue()
	if cell == nil || cell.Size == 0 {
		atomic.StoreInt32(&r.readClosed, 1)
		return 0, io.EOF
	}

	end := cell.Size
	if end > r.mtu {
		end = r.mtu
	}
	n = copy(p, cell.Data[0:end])

	if n < int(end) {
		log.Printf("[DEBUG] fakenet: innerPipe truncated read %d < %d", n, end)
	}

	cell.Size = 0
	cell.Release()

	return n, nil
}

func (r *packetPipe) Write(p []byte) (n int, err error) {
	if atomic.LoadInt32(&r.writeClosed) > 0 {
		return 0, io.ErrClosedPipe
	}
	if len(p) == 0 {
		return 0, nil
	}

	//log.Printf("[DEBUG] fakenet: innerPipe.Write p:%d", len(p))
	if int64(len(p)) > r.mtu {
		//log.Printf("[DEBUG] fakenet: innerPipe truncated write %d > %d", len(p), r.mtu)
		//p = p[0:r.mtu]
		return 0, io.ErrShortBuffer
	}

	cell := r.ringBuffer.Allocate()
	cell.Size = int64(len(p))
	n = copy(cell.Data[0:len(p)], p)
	cell.Enqueue()
	//log.Printf("[DEBUG] fakenet: innerPipe.cellWrite n:%d err:%v", n, err)
	return n, err
}

func (r *packetPipe) Close() (err error) {
	r.closeOnce.Do(func() {
		atomic.StoreInt32(&r.writeClosed, 1)
		// enqueue an empty cell to signal read close
		go func() {
			cell := r.ringBuffer.Allocate()
			cell.Size = 0
			cell.Enqueue()
		}()
	})
	return nil
}
