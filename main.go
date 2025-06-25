package main

import (
	"net/http"
	"time"

	"github.com/thecodinglab/audio/pcm"
	"github.com/thecodinglab/audio/sinks/webrtc"
)

func main() {
	delta, minFreq, maxFreq := 1, 200, 400

	sampler := pcm.NewWaveSampler()
	sampler.Frequency = minFreq
	sampler.SampleRate = 48000
	sampler.Channels = 2
	sampler.Volume = 0.2

	go func() {
		for range time.Tick(10 * time.Millisecond) {
			sampler.Frequency += delta

			if sampler.Frequency <= minFreq {
				sampler.Frequency = minFreq
				delta = -delta
			}

			if sampler.Frequency >= maxFreq {
				sampler.Frequency = maxFreq
				delta = -delta
			}
		}
	}()

	// window := 1 << 16
	// samples := make([]int16, window)
	// n, _ := sampler.Sample(samples)
	//
	// x := make([]complex128, n)
	// for i := range n {
	// 	x[i] = complex(float64(samples[i])/float64(0x7fff), 0)
	// }
	//
	// f := make([]complex128, n)
	// fourier.FFT(x, f)
	//
	// for key, value := range f {
	// 	r, i := real(value), imag(value)
	// 	v := math.Sqrt(r*r + i*i)
	//
	// 	freq := int(math.Round(float64(key) * float64(sampler.SampleRate) / float64(window)))
	// 	fmt.Printf("%4d Hz: %6.4f\n", freq, v)
	// }

	server := webrtc.New(sampler)

	mux := http.NewServeMux()
	mux.Handle("/webrtc", server)
	mux.Handle("/", http.FileServer(http.Dir("./sinks/webrtc")))

	if err := http.ListenAndServe("0.0.0.0:1234", mux); err != nil {
		panic(err)
	}
}
