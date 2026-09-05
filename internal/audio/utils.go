package audio

import (
	"bytes"
	"fmt"
	"io"
	"math"
)

func toReadSeeker(r io.Reader) (io.ReadSeeker, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error while converting to byte slice: %w", err)
	}
	seekableReader := bytes.NewReader(data)
	return seekableReader, nil
}

// converts stereo PCM signal to mono PCM signal
// Stereo signal: [L0 R0 L1 R1 L2 R2 L3 R3]
// Mono signal: [M0 M1 M2 M3] ; where M = (L + R) / 2
// For example, we can have:
//
// Stereo:
// [100, 200, 300, 400]
//
// frames:
//
// [100, 200] → 150
// [300, 400] → 350
//
// result:
//
// [150, 350]
func stereoToMono(samples []int, channels int) ([]float64, error) {
	if channels <= 0 {
		return nil, fmt.Errorf("channels cannot be null")
	}

	if len(samples)%channels != 0 {
		return nil, fmt.Errorf("audio malformed")
	}

	frameCount := len(samples) / channels

	monoAudio := make([]float64, frameCount)

	for frame := 0; frame < frameCount; frame++ {
		var sum float64

		for channel := 0; channel < channels; channel++ {
			idx := frame*channels + channel

			sum += float64(samples[idx])
		}
		monoAudio[frame] = sum / float64(channels)
	}

	return monoAudio, nil
}

// converts everything within a range of [-1.0, 1.0]
func normaliseSamples(samples []float64, bitDepth int) []float64 {
	if bitDepth <= 0 {
		return samples
	}

	scale := math.Pow(
		2,
		float64(bitDepth-1),
	)

	for i := range samples {
		samples[i] /= scale
	}

	return samples
}
