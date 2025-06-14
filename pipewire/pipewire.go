package pipewire

// #cgo pkg-config: libpipewire-0.3
// #include <stdint.h>
// void *audio_setup(const char *name, int sampleRate, int channels, void *userdata);
// void audio_run(void *ctx);
// void audio_quit(void *ctx);
// void audio_close(void *ctx);
import "C"

import (
	"io"
	"math/rand"
	"runtime"
	"sync"
	"unsafe"
)

type Config struct {
	Name       string
	SampleRate int
	Channels   int
}

type Sink struct {
	io.Reader

	cfg Config

	ctx unsafe.Pointer
	wg  sync.WaitGroup
}

func New(reader io.Reader, cfg Config) *Sink {
	sink := &Sink{Reader: reader, cfg: cfg}

	ready := make(chan struct{})

	sink.wg.Add(1)
	go func() {
		defer sink.wg.Done()
		sink.run(ready)
	}()

	<-ready

	return sink
}

func (s *Sink) Config() Config {
	return s.cfg
}

func (s *Sink) Close() {
	s.quit()
	s.wg.Wait()
}

func (s *Sink) run(ready chan struct{}) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	id := registerSink(s)
	defer unregisterSink(id)

	ctx := &context{s.cfg, id}
	userdata := unsafe.Pointer(ctx)

	s.ctx = C.audio_setup(C.CString(s.cfg.Name), C.int(s.cfg.SampleRate), C.int(s.cfg.Channels), userdata)
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
	cfg  Config
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

	n, err := sink.Read(dst)
	if err != nil {
		// TODO: log message?
	}

	return C.size_t(n)
}
