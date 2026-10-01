package audio

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/tosone/minimp3"
)

// MP3Decoder is an AudioDecoder for MP3 audio.
type MP3Decoder struct{}

func NewMP3Decoder() *MP3Decoder {
	return &MP3Decoder{}
}

func (d *MP3Decoder) Decode(r io.Reader) (*AudioBuffer, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error while parsing reader")
	}

	dec, pcm, err := minimp3.DecodeFull(data)
	if err != nil {
		return nil, fmt.Errorf("decode MP3: %w", err)
	}

	if dec.Channels <= 0 {
		return nil, fmt.Errorf("invalid channel count: %d", dec.Channels)
	}

	// minimp3 always decodes to 16-bit signed little-endian PCM.
	intSamples := make([]int, len(pcm)/2)
	for i := range intSamples {
		intSamples[i] = int(int16(binary.LittleEndian.Uint16(pcm[i*2 : i*2+2])))
	}

	samples, err := stereoToMono(intSamples, dec.Channels)
	if err != nil {
		return nil, fmt.Errorf("error converting from stereo to mono")
	}

	samples = normaliseSamples(samples, 16)

	return &AudioBuffer{Samples: samples, SampleRate: dec.SampleRate}, nil
}
