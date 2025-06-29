package sampler

import (
	"io"
	"sync"
)

var _ Sampler = (*Buffer)(nil)

type Buffer struct {
	delegate Sampler

	buf   []int16
	full  bool
	r, w  int
	mutex sync.Mutex
	cond  *sync.Cond
}

func NewBuffer(delegate Sampler) *Buffer {
	format := delegate.Format()
	size := format.SampleRate * format.Channels // buffer 1 second

	sampler := &Buffer{delegate, make([]int16, size), false, 0, 0, sync.Mutex{}, nil}
	sampler.cond = sync.NewCond(&sampler.mutex)

	go func() {
		_, err := sampler.readFrom(delegate)
		sampler.CloseWithError(err)
	}()

	return sampler
}

func (b *Buffer) Format() Format {
	return b.delegate.Format()
}

func (b *Buffer) Sample(buf []int16) (n int, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// TODO
	// if err := r.readErr(true); err != nil {
	// 	return 0, err
	// }
	// if len(p) == 0 {
	// 	return 0, r.readErr(true)
	// }

	for n < len(buf) {
		nr, rerr := b.read(buf[n:])
		n += nr

		if rerr != nil {
			err = rerr
			break
		}
	}

	b.cond.Broadcast()
	return n, err
}

func (b *Buffer) Close() {
	// TODO
}

func (b *Buffer) CloseWithError(err error) {
	// TODO
}

func (b *Buffer) read(buf []int16) (int, error) {
	if b.w == b.r && !b.full {
		// io buffer is currently empty -> fill with zeros
		// TODO instead of filling with zeros, we use the previous values, otherwise there will be artifects
		for i := range buf {
			buf[i] = 0
		}
		return len(buf), nil
	}

	if b.w > b.r {
		// write cursor is further advanced than read cursor -> read until write cursor
		n := min(b.w-b.r, len(buf))
		copy(buf, b.buf[b.r:b.r+n])
		b.r = (b.r + n) % len(b.buf)
		return n, nil
	}

	n := min(len(b.buf)-b.r+b.w, len(buf))

	if b.r+n <= len(b.buf) {
		copy(buf, b.buf[b.r:b.r+n])
	} else {
		p1 := len(b.buf) - b.r
		copy(buf, b.buf[b.r:])
		p2 := n - p1
		copy(buf[p1:], b.buf[0:p2])
	}

	b.r = (b.r + n) % len(b.buf)
	b.full = false

	// TODO return n, r.readErr(true)
	return n, nil
}

func (b *Buffer) readFrom(sampler Sampler) (n int, err error) {
	// zeroReads := 0

	b.mutex.Lock()
	defer b.mutex.Unlock()

	for {
		// if err = r.readErr(true); err != nil {
		// 	return n, err
		// }

		if b.full {
			// buffer is already full -> wait for a read
			b.wait()
			continue
		}

		var buf []int16
		if b.w >= b.r {
			// write cursor is after the reader -> read until end of buffer
			buf = b.buf[b.w:]
		} else {
			// write cursor is before the reader -> read until reader
			buf = b.buf[b.w:b.r]
		}

		// read from origin
		b.mutex.Unlock()
		nr, rerr := sampler.Sample(buf)
		b.mutex.Lock()

		if rerr != nil && rerr != io.EOF {
			// TODO err = r.setErr(rerr, true)
			break
		}

		// if nr == 0 && rerr == nil {
		// 	zeroReads++
		// 	if zeroReads >= 100 {
		// 		err = r.setErr(io.ErrNoProgress, true)
		// 	}
		// 	continue
		// }
		// zeroReads = 0

		b.w += nr
		if b.w == len(b.buf) {
			b.w = 0
		}
		b.full = b.r == b.w && nr > 0
		n += nr
	}

	return n, err
}

func (b *Buffer) wait() {
	// TODO: add timeout?
	b.cond.Wait()
}
