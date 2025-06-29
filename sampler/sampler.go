package sampler

type Format struct {
	SampleRate int
	Channels   int
}

type Sampler interface {
	Format() Format
	Sample(buf []int16) (int, error)
}
