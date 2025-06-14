package pcm

import (
	"errors"
	"io"

	"github.com/smallnest/ringbuffer"
)

type NonBlockingIOSampler struct {
	reader *ringbuffer.RingBuffer
}

func NewNonBlockingIOSampler(reader io.Reader) *NonBlockingIOSampler {
	rb := ringbuffer.New(4 * 1024).SetBlocking(true)
	sampler := &NonBlockingIOSampler{rb}
	go func() {
		_, err := rb.ReadFrom(reader)
		rb.CloseWithError(err)
	}()
	return sampler
}

func (s *NonBlockingIOSampler) Read(buf []byte) (int, error) {
	n, err := s.reader.TryRead(buf)
	if err != nil && !errors.Is(err, ringbuffer.ErrIsEmpty) {
		return n, err
	}

	for i := range len(buf) - n {
		buf[n+i] = 0
	}

	return len(buf), nil
}
