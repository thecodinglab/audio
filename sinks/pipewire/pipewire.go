package pipewire

// #cgo pkg-config: libpipewire-0.3
// #include <stdint.h>
// void *audio_setup(const char *name, int sampleRate, int channels, void *userdata);
// void audio_run(void *ctx);
// void audio_quit(void *ctx);
// void audio_close(void *ctx);
import "C"

import (
	"math/rand"
	"runtime"
	"sync"
	"unsafe"

	"github.com/thecodinglab/audio/pcm"
)

type Sink struct {
	sampler pcm.Sampler

	ctx unsafe.Pointer
	wg  sync.WaitGroup
}

func New(name string, sampler pcm.Sampler) *Sink {
	sink := &Sink{sampler: sampler}

	ready := make(chan struct{})

	sink.wg.Add(1)
	go func() {
		defer sink.wg.Done()
		sink.run(name, ready)
	}()

	<-ready

	return sink
}

func (s *Sink) Close() {
	s.quit()
	s.wg.Wait()
}

func (s *Sink) run(name string, ready chan struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	id := registerSink(s)
	defer unregisterSink(id)

	ctx := &context{id}
	userdata := unsafe.Pointer(ctx)
	format := s.sampler.Format()

	s.ctx = C.audio_setup(C.CString(name), C.int(format.SampleRate), C.int(format.Channels), userdata)
	defer C.audio_close(s.ctx)

	close(ready)

	C.audio_run(s.ctx)
}

func (s *Sink) quit() {
	C.audio_quit(s.ctx)
}

var (
	sinks = map[int64]*Sink{}
	mutex sync.RWMutex
)

func getSink(id int64) *Sink {
	mutex.RLock()
	defer mutex.RUnlock()

	return sinks[id]
}

func registerSink(sink *Sink) int64 {
	mutex.Lock()
	defer mutex.Unlock()

	for {
		id := rand.Int63()
		if _, found := sinks[id]; !found {
			sinks[id] = sink
			return id
		}
	}
}

func unregisterSink(id int64) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(sinks, id)
}

type context struct {
	sink int64
}

//export audio_sample
func audio_sample(buf unsafe.Pointer, size C.size_t, data unsafe.Pointer) C.size_t {
	dst := unsafe.Slice((*byte)(buf), size)
	ctx := (*context)(data)

	sink := getSink(ctx.sink)
	if sink == nil {
		return 0
	}

	n, err := sink.sampler.Read(dst)
	if err != nil {
		// TODO: log message?
	}

	return C.size_t(n)
}
