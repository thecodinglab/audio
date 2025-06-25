package webrtc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"gopkg.in/hraban/opus.v2"

	"github.com/thecodinglab/audio/pcm"
)

type Server struct {
	sampler pcm.Sampler

	tracks map[int64]*webrtc.TrackLocalStaticSample
	mutex  sync.RWMutex
}

func New(sampler pcm.Sampler) *Server {
	server := &Server{sampler, make(map[int64]*webrtc.TrackLocalStaticSample), sync.RWMutex{}}
	go func() {
		if err := server.transfer(); err != nil {
			panic(err)
		}
	}()
	return server
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := s.handleRequest(w, r); err != nil {
		slog.Error(err.Error())
	}
}

func (s *Server) handleRequest(w http.ResponseWriter, r *http.Request) error {
	var offer webrtc.SessionDescription
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return fmt.Errorf("handle request: %w", err)
	}

	conn, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("handle request: %w", err)
	}

	conn.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		fmt.Println("STATE:", pcs)
	})

	ready := make(chan struct{})
	conn.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			close(ready)
		}
	})

	track, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus},
		"audio", "audio",
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("handle request: %w", err)
	}

	sender, err := conn.AddTrack(track)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("handle request: %w", err)
	}

	if err = conn.SetRemoteDescription(offer); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("handle request: %w", err)
	}

	answer, err := conn.CreateAnswer(nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("handle request: %w", err)
	}

	if err = conn.SetLocalDescription(answer); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("handle request: %w", err)
	}

	<-ready
	trackID := s.addTrack(track)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(answer); err != nil {
		return fmt.Errorf("handle request: %w", err)
	}

	go func() {
		defer conn.Close()
		defer s.removeTracks(trackID)

		buf := make([]byte, 1024)
		for {
			if _, _, err := sender.Read(buf); err != nil {
				return
			}
		}
	}()

	return nil
}

func (s *Server) addTrack(track *webrtc.TrackLocalStaticSample) int64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for {
		id := rand.Int63()
		if _, found := s.tracks[id]; !found {
			s.tracks[id] = track
			return id
		}
	}
}

func (s *Server) removeTracks(ids ...int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, id := range ids {
		delete(s.tracks, id)
	}
}

func (s *Server) transfer() error {
	const frameTime = 20 * time.Millisecond

	format := s.sampler.Format()

	enc, err := opus.NewEncoder(format.SampleRate, format.Channels, opus.AppAudio)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(frameTime)
	defer ticker.Stop()

	for range ticker.C {
		numSamples := int(frameTime * time.Duration(format.SampleRate) / time.Second)
		pcm := make([]int16, numSamples*format.Channels)
		if err := binary.Read(s.sampler, binary.LittleEndian, pcm); err != nil {
			return err
		}

		buf := make([]byte, 2*len(pcm))
		n, err := enc.Encode(pcm, buf)
		if err != nil {
			return err
		}

		sample := media.Sample{
			Data:     buf[:n],
			Duration: frameTime,
		}

		var tracksToRemove []int64

		s.mutex.RLock()
		for id, track := range s.tracks {
			if err = track.WriteSample(sample); err != nil {
				tracksToRemove = append(tracksToRemove, id)
			}
		}
		s.mutex.RUnlock()

		if len(tracksToRemove) > 0 {
			s.removeTracks(tracksToRemove...)
		}
	}

	return nil
}
