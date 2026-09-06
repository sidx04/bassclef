# bassclef
An implementation of Shazam's audio fingerprinting.

## Decoder

```
WAV file
   ↓
decode PCM
   ↓
convert channels to mono
   ↓
normalize to float64
   ↓
audio.AudioBuffer
```

```
WAV Decoder ──┐
              │
MP3 Decoder ──┼──> audio.AudioBuffer
              │
Mic Capture ──┘
```

## Frame extraction

An FFT doesn't operate on an entire song. 

Given a sample sequence:
`[0 1 2 3 4 5 6 7 8 9]`, with parameters: `{fft_size = 4 ; hop_size = 2}`, we produce:
```
Frame 0: [0 1 2 3]
Frame 1:     [2 3 4 5]
Frame 2:         [4 5 6 7]
Frame 3:             [6 7 8 9]
```