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

		sampler := pcm.NewWaveSampler()
		sampler.Frequency = 220
		sampler.SampleRate = cfg.SampleRate
		sampler.Channels = cfg.Channels
		sampler.Volume = 0.03

		go func() {
			for range time.Tick(20 * time.Millisecond) {
				sampler.Frequency++
				if sampler.Frequency > 440 {
					sampler.Frequency = 100
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

		sink := pipewire.New(sampler, cfg)
		defer sink.Close()
		<-ctx.Done()
	}()

	wg.Wait()
}
