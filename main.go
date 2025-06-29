package main

import (
	"net/http"
	"time"

	"github.com/thecodinglab/audio/sampler"
	"github.com/thecodinglab/audio/sinks/webrtc"
)

func main() {
	delta, minFreq, maxFreq := 1, 200, 400

	s := sampler.NewWave()
	s.Frequency = minFreq
	s.SampleRate = 48000
	s.Channels = 2
	s.Volume = 0.2

	go func() {
		for range time.Tick(10 * time.Millisecond) {
			s.Frequency += delta

			if s.Frequency <= minFreq {
				s.Frequency = minFreq
				delta = -delta
			}

			if s.Frequency >= maxFreq {
				s.Frequency = maxFreq
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

	buffer := sampler.NewBuffer(s)
	server := webrtc.New(buffer)

	mux := http.NewServeMux()
	mux.Handle("/webrtc", server)
	mux.Handle("/", http.FileServer(http.Dir("./sinks/webrtc")))

	if err := http.ListenAndServe("0.0.0.0:1234", mux); err != nil {
		panic(err)
	}
}
