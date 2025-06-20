package pcm

import (
	"errors"

	"github.com/smallnest/ringbuffer"
)

var _ Sampler = (*NonBlockingSampler)(nil)

type NonBlockingSampler struct {
	delegate Sampler
	reader   *ringbuffer.RingBuffer
}

func NewNonBlockingSampler(delegate Sampler) *NonBlockingSampler {
	rb := ringbuffer.New(4 * 1024).SetBlocking(true)
	sampler := &NonBlockingSampler{delegate, rb}
	go func() {
		_, err := rb.ReadFrom(delegate)
		rb.CloseWithError(err)
	}()
	return sampler
}

func (s *NonBlockingSampler) Format() Format {
	return s.delegate.Format()
}

func (s *NonBlockingSampler) Read(buf []byte) (int, error) {
	n, err := s.reader.TryRead(buf)
	if err != nil && !errors.Is(err, ringbuffer.ErrIsEmpty) {
		return n, err
	}

	for i := range len(buf) - n {
		buf[n+i] = 0
	}

	return len(buf), nil
}
