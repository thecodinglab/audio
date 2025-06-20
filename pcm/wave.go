package pcm

import (
	"encoding/binary"
	"math"
)

var _ Sampler = (*WaveSampler)(nil)

type WaveSampler struct {
	Frequency  int
	SampleRate int
	Channels   int
	Volume     float64

	acc float64
}

func NewWaveSampler() *WaveSampler {
	return &WaveSampler{220, 44100, 1, 2, 0}
}

func (s *WaveSampler) Format() Format {
	return Format{
		SampleRate: s.SampleRate,
		Channels:   s.Channels,
	}
}

const (
	twoPI     = math.Pi + math.Pi
	sizeInt16 = 2
)

func (s *WaveSampler) Read(buf []byte) (int, error) {
	n := 0
	frames := len(buf) / (sizeInt16 * s.Channels)

	for i := range frames {
		s.acc += twoPI * float64(s.Frequency) / float64(s.SampleRate)
		for s.acc > twoPI {
			s.acc -= twoPI
		}

		val := int16(math.Sin(s.acc) * 0x7fff * s.Volume)
		for c := range s.Channels {
			idx := sizeInt16 * (s.Channels*i + c)
			binary.LittleEndian.PutUint16(buf[idx:], uint16(val))
			n += 2
		}
	}

	return n, nil
}
