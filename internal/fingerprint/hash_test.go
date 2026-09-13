package fingerprint

import "testing"

func TestHashKnownValue(t *testing.T) {
	got := Hash(1, 2, 3)
	want := uint64(1)<<26 | uint64(2)<<14 | uint64(3)

	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestHashDistinctInputsDistinctOutputs(t *testing.T) {
	a := Hash(100, 200, 5)
	b := Hash(100, 200, 6)
	c := Hash(100, 201, 5)
	d := Hash(101, 200, 5)

	all := []uint64{a, b, c, d}
	for i := range all {
		for j := range all {
			if i == j {
				continue
			}
			if all[i] == all[j] {
				t.Errorf("hash collision: index %d and %d both %v", i, j, all[i])
			}
		}
	}
}

func TestHashMasksOutOfRangeFields(t *testing.T) {
	// f1 has only 12 bits of room; a value one bit wider must not bleed
	// into f2's field.
	got := Hash(1<<12, 0, 0)
	want := uint64(0)

	if got != want {
		t.Errorf("got %v, want %v (overflow bit should be masked away)", got, want)
	}
}
