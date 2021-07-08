package slowrdr

import (
	"io"
	"time"
)

var _ io.Reader = &slowReader{}

type slowReader struct {
	r      io.Reader // the reader to read from
	limit  int64     // max bytes per second
	second int64     // the current second
	read   int64     // bytes read this second
}

func NewSlowReader(r io.Reader, maxBytesPerSecond int64) io.Reader {
	return &slowReader{r: r, limit: maxBytesPerSecond}
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	now := time.Now().Unix()
	if now > r.second {
		r.second = now
		r.read = 0
	}
	pSize := int64(len(p))
	if r.read+pSize > r.limit {
		maxRead := r.limit - r.read + pSize
		if maxRead <= 0 {
			time.Sleep(time.Until(time.Unix(now+1, 0)))
			return r.Read(p)
		}
		if maxRead > pSize {
			maxRead = pSize
		}
		r.read += maxRead
		return r.r.Read(p[0:maxRead])
	}
	r.read += pSize
	return r.r.Read(p)
}
