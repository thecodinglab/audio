package sampler

import (
	"math"
)

var _ Sampler = (*Wave)(nil)

type Wave struct {
	Frequency  int
	SampleRate int
	Channels   int
	Volume     float64

	acc float64
}

func NewWave() *Wave {
	return &Wave{220, 44100, 1, 2, 0}
}

func (s *Wave) Format() Format {
	return Format{
		SampleRate: s.SampleRate,
		Channels:   s.Channels,
	}
}

const (
	twoPI     = math.Pi + math.Pi
	sizeInt16 = 2
)

func (s *Wave) Sample(samples []int16) (int, error) {
	n := 0
	frames := len(samples) / s.Channels

	for i := range frames {
		s.acc += twoPI * float64(s.Frequency) / float64(s.SampleRate)
		for s.acc > twoPI {
			s.acc -= twoPI
		}

		val := int16(math.Sin(s.acc) * 0x7fff * s.Volume)
		for c := range s.Channels {
			idx := i*s.Channels + c
			samples[idx] = val
			n++
		}
	}

	return n, nil
}
