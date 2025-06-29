package pipewire

// #cgo pkg-config: libpipewire-0.3
// #include <stdint.h>
// void *audio_setup(const char *name, int sampleRate, int channels, uint64_t userdata);
// void audio_run(void *ctx);
// void audio_quit(void *ctx);
// void audio_close(void *ctx);
import "C"

import (
	"encoding/binary"
	"math/rand"
	"runtime"
	"sync"
	"unsafe"

	"github.com/thecodinglab/audio/sampler"
)

type Sink struct {
	sampler sampler.Sampler

	ctx unsafe.Pointer
	wg  sync.WaitGroup
}

func New(name string, sampler sampler.Sampler) *Sink {
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

	format := s.sampler.Format()

	s.ctx = C.audio_setup(C.CString(name), C.int(format.SampleRate), C.int(format.Channels), C.uint64_t(id))
	defer C.audio_close(s.ctx)

	close(ready)

	C.audio_run(s.ctx)
}

func (s *Sink) quit() {
	C.audio_quit(s.ctx)
}

var (
	sinks = map[uint64]*Sink{}
	mutex sync.RWMutex
)

func getSink(id uint64) *Sink {
	mutex.RLock()
	defer mutex.RUnlock()

	return sinks[id]
}

func registerSink(sink *Sink) uint64 {
	mutex.Lock()
	defer mutex.Unlock()

	for {
		id := rand.Uint64()
		if _, found := sinks[id]; !found {
			sinks[id] = sink
			return id
		}
	}
}

func unregisterSink(id uint64) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(sinks, id)
}

//export audio_sample
func audio_sample(buf unsafe.Pointer, size C.size_t, id C.uint64_t) C.size_t {
	dst := unsafe.Slice((*byte)(buf), size)

	sink := getSink(uint64(id))
	if sink == nil {
		return 0
	}

	samples := make([]int16, size/2)
	n, err := sink.sampler.Sample(samples)
	if err != nil {
		// TODO: log message?
	}

	n, err = binary.Encode(dst, binary.LittleEndian, samples[:n])
	if err != nil {
		// TODO: log message?
	}

	return C.size_t(n)
}
