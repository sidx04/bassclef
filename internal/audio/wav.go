package audio

import (
	"fmt"
	"io"

	"github.com/go-audio/wav"
)

func NewWAVDecoder() *WAVDecoder {
	return &WAVDecoder{}
}

func (d *WAVDecoder) Decode(r io.Reader) (*AudioBuffer, error) {
	sr, err := toReadSeeker(r)
	if err != nil {
		return nil, fmt.Errorf("error while parsing reader")
	}

	decoder := wav.NewDecoder(sr)

	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV file")
	}

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("decode WAV: %w", err)
	}

	if buf == nil {
		return nil, fmt.Errorf("decoded WAV buffer is empty")
	}

	if buf.Format == nil {
		return nil, fmt.Errorf("WAV format is missing")
	}

	if buf.Format.NumChannels <= 0 {
		return nil, fmt.Errorf("invalid channel count: %d", buf.Format.NumChannels)
	}

	samples, err := stereoToMono(buf.Data, buf.Format.NumChannels)
	if err != nil {
		return nil, fmt.Errorf("error converting from stereo to mono")
	}

	samples = normaliseSamples(samples, buf.SourceBitDepth)

	return &AudioBuffer{Samples: samples, SampleRate: buf.Format.SampleRate}, nil
}
