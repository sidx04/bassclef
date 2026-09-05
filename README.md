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