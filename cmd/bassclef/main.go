package main

import "github.com/alecthomas/kong"

var CLI struct {
	Info        InfoCmd        `cmd:"" help:"Display information about an audio file."`
	Spectrogram SpectrogramCmd `cmd:"" help:"Generate an audio spectrogram."`
	// Peaks       PeaksCmd       `cmd:"" help:"Detect spectral peaks."`
	// Fingerprint FingerprintCmd `cmd:"" help:"Generate audio fingerprints."`
	// Ingest      IngestCmd      `cmd:"" help:"Add a song to the catalog."`
	// Recognize   RecognizeCmd   `cmd:"" help:"Recognize an audio sample."`
}

func main() {
	ctx := kong.Parse(&CLI)

	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
