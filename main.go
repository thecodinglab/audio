package main

import (
	"time"

	"github.com/thecodinglab/audio/pipewire"
)

func main() {
	sink := pipewire.New()
	defer sink.Close()

	sink.Ready()
	time.Sleep(4 * time.Second)
}
