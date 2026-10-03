package dsp

import "math"

// NormalizeGain scales samples in place so the loudest sample reaches
// full scale (|sample| == 1.0); silence (all-zero input) is left
// unchanged. Mic recordings vary widely in loudness by distance and
// volume, and MinMagnitude is an absolute threshold — without this, a
// quiet recording of a clear song can produce zero peaks while a louder
// one of the same song clears the threshold easily. Normalizing every
// buffer the same way keeps MinMagnitude comparisons meaningful
// regardless of source loudness.
func NormalizeGain(samples []float64) {
	peak := 0.0
	for _, s := range samples {
		if abs := math.Abs(s); abs > peak {
			peak = abs
		}
	}

	if peak == 0 {
		return
	}

	scale := 1.0 / peak
	for i := range samples {
		samples[i] *= scale
	}
}
