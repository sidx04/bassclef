package fingerprint

const (
	freqBits  = 12
	deltaBits = 14

	freqMask  = 1<<freqBits - 1
	deltaMask = 1<<deltaBits - 1
)

// Hash packs an anchor/target frequency bin pair and their time delta into
// a compact, deterministic key: [f1:12 bits][f2:12 bits][dt:14 bits]. Not
// cryptographic. Each field is masked to its bit width before packing, so
// a value wider than its field cannot bleed into a neighboring one.
func Hash(f1, f2, dt uint32) uint64 {
	return uint64(f1&freqMask)<<(freqBits+deltaBits) |
		uint64(f2&freqMask)<<deltaBits |
		uint64(dt&deltaMask)
}
