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

## Frame extraction and Hann Window

An FFT doesn't operate on an entire song. 

Given a sample sequence:
`[0 1 2 3 4 5 6 7 8 9]`, with parameters: `{fft_size = 4 ; hop_size = 2}`, we produce:
```
Frame 0: [0 1 2 3]
Frame 1:     [2 3 4 5]
Frame 2:         [4 5 6 7]
Frame 3:             [6 7 8 9]
```

The FFT assumes the signal repeats periodically. Simply cutting arbitrary chunks out of an audio signal creates discontinuities at the boundaries. 

Take a look at these: 
![sampled-signal](https://community.sw.siemens.com/servlet/rtaImage?eid=ka6Vb000001pPv3&feoid=00N4O000006Yxpf&refid=0EM4O00000113sj)
![normalisation-signal](https://community.sw.siemens.com/servlet/rtaImage?eid=ka6Vb000001pPv3&feoid=00N4O000006Yxpf&refid=0EM4O00000113sk)

Notice the problem; there is some sort of signal leakage due to the non-periodic smaple. Random data has spectral leakage due to the abrupt cutoff at the beginning and end of the time block. It is non-periodic. There is no way to ensure that the captured random signal is periodic by varying the measurement time. A random signal is composed of many different frequencies, and even if the acquisition time was adjusted to capture some frequencies without creating leakage, other frequencies would be non-periodic.

Therefore, we use the [**Hann Window**](hann-window). The Hann window smoothly reduces the frame edges toward zero.

![hann-1](https://community.sw.siemens.com/servlet/rtaImage?eid=ka6Vb000001pPv3&feoid=00N4O000006Yxpf&refid=0EM4O00000113sm)

Mathematically, we represent the Hann Window as follows:

$$
w(n)=\frac{1}{2}\left(1-cos\left({\frac{2 \pi n}{N-1}} \right) \right)
$$
where, $N$ is the size of the window. We iterate from $n=0$ to $N-1$. 

