package pipewire

// #cgo pkg-config: libpipewire-0.3
// #include <stdint.h>
// void *audio_setup(const char *name, int sampleRate, int channels, void *userdata);
// void audio_run(void *ctx);
// void audio_quit(void *ctx);
// void audio_close(void *ctx);
import "C"

import (
	"math"
	"math/rand"
	"runtime"
	"sync"
	"unsafe"
)

const (
	twoPI      = math.Pi + math.Pi
	sampleRate = 44100
	channels   = 2
)

type Sink struct {
	ctx   unsafe.Pointer
	ready chan struct{}
	wg    sync.WaitGroup
}

func New(name string) *Sink {
	sink := &Sink{ready: make(chan struct{})}

	sink.wg.Add(1)
	go func() {
		defer sink.wg.Done()
		sink.run(name)
	}()

	return sink
}

func (s *Sink) Ready() {
	<-s.ready
}

func (s *Sink) Close() {
	s.quit()
	s.wg.Wait()
}

func (s *Sink) run(name string) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ctx := &context{freq: rand.Float64() * 440}
	userdata := unsafe.Pointer(ctx)

	s.ctx = C.audio_setup(C.CString(name), sampleRate, channels, userdata)
	defer C.audio_close(s.ctx)

	close(s.ready)

	C.audio_run(s.ctx)
}

func (s *Sink) quit() {
	C.audio_quit(s.ctx)
}

type context struct {
	freq float64
	acc  float64
}

//export audio_sample
func audio_sample(buf *C.int16_t, size C.size_t, data unsafe.Pointer) {
	dst := unsafe.Slice(buf, size)
	userdata := (*context)(data)

	stride := 2 /* sizeof(int16_t) */ * channels
	n_frames := int(size) / stride

	for i := range n_frames {
		userdata.acc += twoPI * userdata.freq / sampleRate
		for userdata.acc > twoPI {
			userdata.acc -= twoPI
		}

		val := int16(math.Sin(userdata.acc) * 0.03 * 32767.0)
		for c := range 2 {
			dst[2*i+c] = C.int16_t(val)
		}
	}
}
