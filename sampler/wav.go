package sampler

import (
	"encoding/binary"
	"errors"
	"io"
)

var _ Sampler = (*WaveSampler)(nil)

type WaveRIFFHeader struct {
	FileSize uint32
}

type WaveDataFormatHeader struct {
	BlockSize      uint32
	AudioFormat    uint16
	Channels       uint16
	SampleRate     uint32
	BytesPerSecond uint32
	BytesPerBlock  uint16
	BitsPerSample  uint16
}

type WaveSampledDataHeader struct {
	Size uint32
}

const WaveHeaderSize = 44

type WaveSampler struct {
	riff   WaveRIFFHeader
	format WaveDataFormatHeader
	data   WaveSampledDataHeader

	reader io.Reader
}

func NewWAV(reader io.Reader) (*WaveSampler, error) {
	riff, err := readRIFFHeader(reader)
	if err != nil {
		return nil, err
	}

	format, err := readDataFormatHeader(reader)
	if err != nil {
		return nil, err
	}

	data, err := readSampledDataHeader(reader)
	if err != nil {
		return nil, err
	}

	return &WaveSampler{riff, format, data, reader}, nil
}

func (s *WaveSampler) Format() Format {
	return Format{
		SampleRate: int(s.format.SampleRate),
		Channels:   int(s.format.Channels),
	}
}

func (s *WaveSampler) Sample(buf []int16) (int, error) {
	data := make([]byte, len(buf)*2)
	n, err := s.reader.Read(data)
	if n > 0 {
		n, err = binary.Decode(data[:n], binary.LittleEndian, buf)
		return n / 2, err
	}
	return 0, err
}

var ErrInvalidWAV = errors.New("invalid wav")

func readRIFFHeader(reader io.Reader) (WaveRIFFHeader, error) {
	var b [12]byte
	buf := b[:]

	if _, err := io.ReadAtLeast(reader, buf[:], len(buf)); err != nil {
		return WaveRIFFHeader{}, err
	}

	header := WaveRIFFHeader{}

	if string(buf[0:4]) != "RIFF" {
		return WaveRIFFHeader{}, ErrInvalidWAV
	}
	buf = buf[4:]

	header.FileSize = binary.LittleEndian.Uint32(buf)
	buf = buf[4:]

	if string(buf[:4]) != "WAVE" {
		return WaveRIFFHeader{}, ErrInvalidWAV
	}
	buf = buf[4:]

	return header, nil
}

func readDataFormatHeader(reader io.Reader) (WaveDataFormatHeader, error) {
	var b [24]byte
	buf := b[:]

	if _, err := io.ReadAtLeast(reader, buf[:], len(buf)); err != nil {
		return WaveDataFormatHeader{}, err
	}

	header := WaveDataFormatHeader{}

	if string(buf[:4]) != "fmt " {
		return WaveDataFormatHeader{}, ErrInvalidWAV
	}
	buf = buf[4:]

	header.BlockSize = binary.LittleEndian.Uint32(buf[:4])
	buf = buf[4:]

	header.AudioFormat = binary.LittleEndian.Uint16(buf[:2])
	buf = buf[2:]

	header.Channels = binary.LittleEndian.Uint16(buf[:2])
	buf = buf[2:]

	header.SampleRate = binary.LittleEndian.Uint32(buf[:4])
	buf = buf[4:]

	header.BytesPerSecond = binary.LittleEndian.Uint32(buf[:4])
	buf = buf[4:]

	header.BytesPerBlock = binary.LittleEndian.Uint16(buf[:2])
	buf = buf[2:]

	header.BitsPerSample = binary.LittleEndian.Uint16(buf[:2])
	buf = buf[2:]

	return header, nil
}

func readSampledDataHeader(reader io.Reader) (WaveSampledDataHeader, error) {
	var b [8]byte
	buf := b[:]

	if _, err := io.ReadAtLeast(reader, buf[:], len(buf)); err != nil {
		return WaveSampledDataHeader{}, err
	}

	header := WaveSampledDataHeader{}

	if string(buf[:4]) != "data" {
		return WaveSampledDataHeader{}, ErrInvalidWAV
	}
	buf = buf[4:]

	header.Size = binary.LittleEndian.Uint32(buf[:4])
	buf = buf[4:]

	return header, nil
}
