package pcm

import "io"

type Format struct {
	SampleRate int
	Channels   int
}

type Sampler interface {
	io.Reader

	Format() Format
}
