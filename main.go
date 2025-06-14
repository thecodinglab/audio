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
	wg.Add(2)

	go func() {
		defer wg.Done()

		sink := pipewire.New("ananas")
		defer sink.Close()

		sink.Ready()
		<-ctx.Done()
	}()

	go func() {
		defer wg.Done()

		sink := pipewire.New("banane")
		defer sink.Close()

		sink.Ready()
		<-ctx.Done()
	}()

	wg.Wait()
}
