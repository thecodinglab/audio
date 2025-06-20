package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/thecodinglab/audio/pcm"
	"github.com/thecodinglab/audio/sinks/pipewire"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		delta, minFreq, maxFreq := 1, 200, 400

		sampler := pcm.NewWaveSampler()
		sampler.Frequency = minFreq
		sampler.SampleRate = 44100
		sampler.Channels = 2
		sampler.Volume = 0.03

		go func() {
			for range time.Tick(time.Millisecond) {
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

		sink := pipewire.New("ananas", sampler)
		defer sink.Close()
		<-ctx.Done()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		sampler := pcm.NewWaveSampler()
		sampler.SampleRate = 44100 / 2
		sampler.Channels = 1
		sampler.Volume = 0.03

		unblocked := pcm.NewNonBlockingSampler(sampler)

		sink := pipewire.New("banane", unblocked)
		defer sink.Close()
		<-ctx.Done()
	}()

	wg.Wait()
}
