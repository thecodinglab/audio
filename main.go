package main

import (
	"context"
	"os"
	"os/signal"
	"sync"

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
			SampleRate: 44100,
			Channels:   2,
		}

		sink := pipewire.New("ananas", cfg)
		defer sink.Close()

		sink.Ready()
		<-ctx.Done()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		cfg := pipewire.Config{
			SampleRate: 44100 / 2,
			Channels:   1,
		}

		sink := pipewire.New("banane", cfg)
		defer sink.Close()

		sink.Ready()
		<-ctx.Done()
	}()

	wg.Wait()
}
