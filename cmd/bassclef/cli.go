package main

import (
	"fmt"
	"log"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/dsp"
	"github.com/sidx04/bassclef/internal/fingerprint"
)

type InfoCmd struct {
	File string `arg:"" name:"file" help:"Path to the audio file." type:"path"`
}

func (cmd *InfoCmd) Run() error {
	buf, err := audio.LoadWAV(cmd.File)
	if err != nil {
		log.Fatalf("failed to load audio: %v", err)
	}

	duration := float64(len(buf.Samples)) / float64(buf.SampleRate)

	fmt.Printf("Sample rate: %d Hz\n", buf.SampleRate)
	fmt.Printf("Samples: %d\n", len(buf.Samples))
	fmt.Printf("Duration: %.2f seconds\n", duration)

	return nil
}

type SpectrogramCmd struct {
	File string `arg:"" name:"file" help:"Path to the audio file." type:"path"`
}

func (cmd *SpectrogramCmd) Run() error {
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

	fmt.Printf("Frames: %d\n", len(spectrogram))
	fmt.Printf("Bins per frame: %d\n", cfg.FFTSize)

	return nil
}

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
