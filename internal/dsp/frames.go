package dsp

import "fmt"

// Split the song into frames for FFT. Check [this section] of the documentation.
//
// [this section]: https://github.com/sidx04/bassclef.git/README.md##Frame-Extraction
func SplitFrames(samples []float64, frameSize int, hopSize int) ([][]float64, error) {
	if frameSize <= 0 {
		return nil, fmt.Errorf("frame size cannot be less than 0: %d", frameSize)
	}

	if hopSize <= 0 {
		return nil, fmt.Errorf("hop size cannot be less than 0: %d", frameSize)
	}

	if len(samples) < frameSize {
		return nil, nil
	}

	frameCount := (len(samples)-frameSize)/hopSize + 1

	frames := make([][]float64, 0, frameCount)

	for start := 0; start+frameSize <= len(samples); start += hopSize {
		frame := make([]float64, frameSize)

		copy(frame, samples[start:start+frameSize])

		frames = append(frames, frame)
	}

	return frames, nil
}
