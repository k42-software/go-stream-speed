package fakenet

import (
	"bytes"
	"github.com/k42-software/go-stream-speed/pkg/chnkcpy"
	"io"
	"sync"
	"sync/atomic"
)

var _ io.ReadWriteCloser = &streamPipe{}

// FIFO stream based NetworkPipe backed by a ring buffer
type streamPipe struct {
	innerPipe   NetworkPipe // This is where we write the packets to
	readPending []byte
	readLock    *sync.Mutex
	writeLock   *sync.Mutex

	// close management
	writeClosed int32
	readClosed  int32
	closeOnce   sync.Once
}

// Stream based NetworkPipe backed by a ring buffer
func NewStreamPipe(mtu int64) (NetworkPipe, error) {
	innerPipe, err := NewPacketPipe(mtu)
	if innerPipe == nil {
		return nil, err
	}
	return &streamPipe{
		innerPipe:   innerPipe,
		readPending: make([]byte, 0, innerPipe.MTU()),
		readLock:    &sync.Mutex{},
		writeLock:   &sync.Mutex{},
	}, err
}

func (r *streamPipe) MTU() int64 {
	return r.innerPipe.MTU()
}

// Each read can consume at most one virtual packet, but reads which present
// a buffer smaller than the MTU will be spread over a single packet until it
// is consumed. All reads are handled sequentially.
func (r *streamPipe) Read(p []byte) (n int, err error) {
	// stream: all reads have to be sequential
	r.readLock.Lock()
	defer r.readLock.Unlock()

	if atomic.LoadInt32(&r.readClosed) > 0 {
		return 0, io.EOF
	}
	//log.Printf("[DEBUG] fakenet: streamPipe.Read p:%d", len(p))

	// if there are pending bytes, consume them before going to the buffer
	if len(r.readPending) > 0 {
		n = copy(p, r.readPending)
		r.readPending = r.readPending[n:]
		return n, nil
	}

	mtu := r.MTU()
	if int64(len(r.readPending)) < mtu {
		if int64(cap(r.readPending)) >= mtu {
			r.readPending = r.readPending[0:mtu]
		} else {
			r.readPending = make([]byte, mtu)
		}
	}

	packetSize := 0
	packetSize, err = r.innerPipe.Read(r.readPending)
	if err != nil {
		if err == io.EOF {
			atomic.StoreInt32(&r.readClosed, 1)
		}
		return 0, err
	}

	end := packetSize
	if int64(end) > mtu {
		end = int(mtu)
	}
	n = copy(p, r.readPending[0:end])

	if n < end {
		// if we consumed less than the whole packet, then we keep the rest for the
		// next read - this is what makes us a stream ;)
		r.readPending = r.readPending[n:end]
	} else {
		// clear the read packet
		r.readPending = r.readPending[:0]
	}

	return n, nil
}

// Each write smaller than the MTU will result in one virtual packet. Each
// writer larger than the MTU will be fragmented/chunked into MTU sized
// virtual packets. All writes are handled sequentially.
func (r *streamPipe) Write(p []byte) (n int, err error) {
	// stream: all writes have to be sequential
	r.writeLock.Lock()
	defer r.writeLock.Unlock()

	if atomic.LoadInt32(&r.writeClosed) > 0 {
		return 0, io.ErrClosedPipe
	}
	if len(p) == 0 {
		return 0, nil
	}

	//log.Printf("[DEBUG] fakenet: streamPipe.Write p:%d", len(p))
	if int64(len(p)) <= r.MTU() {
		return r.innerPipe.Write(p)
	}

	// This is intended to split the given data up into chunks
	// of the size CellSize and write them directly in to the
	// ring buffer (bypassing the mutex that we already hold).
	return chnkcpy.ChunkedCopy(
		chnkcpy.WriterFunc(func(p []byte) (n int, err error) {
			return r.innerPipe.Write(p)
		}),
		bytes.NewReader(p),
		int(r.MTU()),
	)
}

func (r *streamPipe) Close() (err error) {
	r.closeOnce.Do(func() {
		r.writeLock.Lock()
		defer r.writeLock.Unlock()
		atomic.StoreInt32(&r.writeClosed, 1)
		err = r.innerPipe.Close()
	})
	return err
}
