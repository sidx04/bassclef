package audio

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
)

// captureSampleRate must match the sample rate WAV fixtures are ingested
// at (44100 Hz). Landmark time offsets are a sample-count-based unit,
// not a real-time one, so a mismatched capture rate
// would silently misalign matching against the catalog.
const captureSampleRate = 44100 // in hertz

// Capture records duration of audio from the default input device and
// returns it as a normalized mono AudioBuffer.
func Capture(duration time.Duration) (*AudioBuffer, error) {
	if duration <= 0 {
		return nil, fmt.Errorf("duration must be positive: %s", duration)
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(string) {})
	if err != nil {
		return nil, fmt.Errorf("failed to init audio context: %w", err)
	}
	defer ctx.Free()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = captureSampleRate

	var mu sync.Mutex
	var raw []byte

	onData := func(_, input []byte, _ uint32) {
		mu.Lock()
		raw = append(raw, input...)
		mu.Unlock()
	}

	device, err := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{Data: onData})
	if err != nil {
		return nil, fmt.Errorf("failed to init capture device: %w", err)
	}
	defer device.Uninit()

	if err := device.Start(); err != nil {
		return nil, fmt.Errorf("failed to start capture: %w", err)
	}

	time.Sleep(duration)

	if err := device.Stop(); err != nil {
		return nil, fmt.Errorf("failed to stop capture: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()

	samples := make([]float64, len(raw)/2)
	for i := range samples {
		s := int16(binary.LittleEndian.Uint16(raw[i*2 : i*2+2]))
		samples[i] = float64(s)
	}

	samples = normaliseSamples(samples, 16)

	return &AudioBuffer{Samples: samples, SampleRate: captureSampleRate}, nil
}
