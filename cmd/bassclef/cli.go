package main

import (
	"fmt"
	"log"

	"github.com/sidx04/bassclef/internal/audio"
	"github.com/sidx04/bassclef/internal/dsp"
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
