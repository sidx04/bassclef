package main

import (
	"fmt"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/dsp"
	"github.com/sidx04/bassclef/internal/fingerprint"
)

type PeaksCmd struct {
	File         string  `arg:"" name:"file" help:"Path to the audio file." type:"path"`
	FreqWindow   int     `name:"freq-window" default:"3" help:"Bins each side of a candidate peak."`
	TimeWindow   int     `name:"time-window" default:"3" help:"Frames each side of a candidate peak."`
	MinMagnitude float64 `name:"min-magnitude" default:"0" help:"Reject candidates below this magnitude."`
}

func (cmd *PeaksCmd) Run() error {
	buf, err := audio.LoadWAV(cmd.File)
	if err != nil {
		return fmt.Errorf("failed to load audio: %w", err)
	}

	cfg := dsp.DefaultConfig()

	frames, err := dsp.SplitFrames(buf.Samples, cfg.FFTSize, cfg.HopSize)
	if err != nil {
		return fmt.Errorf("failed to split frames: %w", err)
	}

	window, err := dsp.HannWindow(cfg.FFTSize)
	if err != nil {
		return fmt.Errorf("failed to build window: %w", err)
	}

	fft := dsp.NewFFT(cfg.FFTSize)

	spectrogram, err := dsp.Spectrogram(frames, window, fft)
	if err != nil {
		return fmt.Errorf("failed to compute spectrogram: %w", err)
	}

	peakCfg := fingerprint.PeakConfig{
		FreqWindow:   cmd.FreqWindow,
		TimeWindow:   cmd.TimeWindow,
		MinMagnitude: cmd.MinMagnitude,
	}

	peaks, err := fingerprint.FindPeaks(spectrogram, peakCfg)
	if err != nil {
		return fmt.Errorf("failed to find peaks: %w", err)
	}

	fmt.Printf("Peaks: %d\n", len(peaks))

	for _, p := range peaks {
		hz := fingerprint.BinToHz(p.Frequency, buf.SampleRate, cfg.FFTSize)
		fmt.Printf("time=%d freq=%.1fHz magnitude=%.4f\n", p.Time, hz, p.Magnitude)
	}

	return nil
}
