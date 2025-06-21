package main

import (
	"fmt"
	"math"

	"github.com/thecodinglab/audio/fourier"
	"github.com/thecodinglab/audio/pcm"
)

func main() {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// defer cancel()
	//
	// wg := sync.WaitGroup{}
	//
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	//
	// 	delta, minFreq, maxFreq := 1, 200, 400
	//
	// 	sampler := pcm.NewWaveSampler()
	// 	sampler.Frequency = minFreq
	// 	sampler.SampleRate = 44100
	// 	sampler.Channels = 2
	// 	sampler.Volume = 0.03
	//
	// 	go func() {
	// 		for range time.Tick(time.Millisecond) {
	// 			sampler.Frequency += delta
	//
	// 			if sampler.Frequency <= minFreq {
	// 				sampler.Frequency = minFreq
	// 				delta = -delta
	// 			}
	//
	// 			if sampler.Frequency >= maxFreq {
	// 				sampler.Frequency = maxFreq
	// 				delta = -delta
	// 			}
	// 		}
	// 	}()
	//
	// 	sink := pipewire.New("ananas", sampler)
	// 	defer sink.Close()
	// 	<-ctx.Done()
	// }()
	//
	// wg.Add(1)
	// go func() {
	// 	defer wg.Done()
	//
	// 	sampler := pcm.NewWaveSampler()
	// 	sampler.SampleRate = 44100 / 2
	// 	sampler.Channels = 1
	// 	sampler.Volume = 0.03
	//
	// 	unblocked := pcm.NewNonBlockingSampler(sampler)
	//
	// 	sink := pipewire.New("banane", unblocked)
	// 	defer sink.Close()
	// 	<-ctx.Done()
	// }()
	//
	// wg.Wait()

	sampler := pcm.NewWaveSampler()
	sampler.SampleRate = 44100
	sampler.Frequency = 220
	sampler.Channels = 1
	sampler.Volume = 1

	window := 1 << 13

	samples := make([]int16, window)
	n, _ := sampler.Sample(samples)

	x := make([]complex128, n)
	for i := range n {
		x[i] = complex(float64(samples[i])/float64(0x7fff), 0)
	}

	f := make([]complex128, n)
	fourier.DFT(x, f)

	for key, value := range f {
		r, i := real(value), imag(value)
		v := math.Sqrt(r*r + i*i)

		freq := float64(key) * float64(sampler.SampleRate) / float64(window)
		fmt.Printf("%4d Hz: %6.4f\n", int(freq), v)
	}
}
