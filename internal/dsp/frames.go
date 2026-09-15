package dsp

import "fmt"

// SplitFrames splits samples into overlapping fixed-size frames for FFT
// processing. Each frame holds frameSize samples; consecutive frames start
// hopSize samples apart, so hopSize < frameSize gives overlapping frames
// and hopSize == frameSize gives non-overlapping frames.
//
// Returns (nil, nil) if len(samples) < frameSize: not enough data for one
// frame. Returns an error if frameSize or hopSize is not positive.
//
// Check [this section] of the documentation.
//
// [this section]: https://github.com/sidx04/bassclef/blob/main/README.md#frame-extraction
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

	frames := extractSplitFrames(samples, frameSize, hopSize)

	return frames, nil
}

func extractSplitFrames(samples []float64, frameSize int, hopSize int) [][]float64 {
	frameCount := (len(samples)-frameSize)/hopSize + 1

	frames := make([][]float64, 0, frameCount)

	for start := 0; start+frameSize <= len(samples); start += hopSize {
		frame := make([]float64, frameSize)

		copy(frame, samples[start:start+frameSize])

		frames = append(frames, frame)
	}
	return frames
}
