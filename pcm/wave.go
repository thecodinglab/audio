package pcm

import (
	"encoding/binary"
	"math"
)

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

const twoPI = math.Pi + math.Pi

func (s *WaveSampler) Read(buf []byte) (int, error) {
	n := 0
	frames := len(buf) / (2 /* int16 */ * s.Channels)

	for i := range frames {
		s.acc += twoPI * float64(s.Frequency) / float64(s.SampleRate)
		for s.acc > twoPI {
			s.acc -= twoPI
		}

		val := int16(math.Sin(s.acc) * 32767.0 * s.Volume)
		for c := range s.Channels {
			idx := 2 * (s.Channels*i + c)
			binary.LittleEndian.PutUint16(buf[idx:], uint16(val))
			n += 2
		}
	}

	return n, nil
}
