package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/thecodinglab/audio/pcm"
	"github.com/thecodinglab/audio/pipewire"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		cfg := pipewire.Config{
			Name:       "ananas",
			SampleRate: 44100,
			Channels:   2,
		}

		delta, minFreq, maxFreq := 1, 200, 400

		sampler := pcm.NewWaveSampler()
		sampler.Frequency = minFreq
		sampler.SampleRate = cfg.SampleRate
		sampler.Channels = cfg.Channels
		sampler.Volume = 0.03

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

		sink := pipewire.New(sampler, cfg)
		defer sink.Close()
		<-ctx.Done()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		cfg := pipewire.Config{
			Name:       "banane",
			SampleRate: 44100 / 2,
			Channels:   1,
		}

		sampler := pcm.NewWaveSampler()
		sampler.SampleRate = cfg.SampleRate
		sampler.Channels = cfg.Channels
		sampler.Volume = 0.03

		unblocked := pcm.NewNonBlockingIOSampler(sampler)

		sink := pipewire.New(unblocked, cfg)
		defer sink.Close()
		<-ctx.Done()
	}()

	wg.Wait()
}
