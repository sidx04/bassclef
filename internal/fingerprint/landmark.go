package fingerprint

import (
	"fmt"
	"sort"
)

// GenerateLandmarks pairs each peak, as an anchor, with the nearest-in-time
// peaks that follow it within cfg's delta window, up to cfg.Fanout pairs
// per anchor. peaks need not be time-sorted; a sorted copy is made
// internally.
func GenerateLandmarks(peaks []Peak, cfg LandmarkConfig) ([]Landmark, error) {
	if cfg.MinDeltaFrames < 0 {
		return nil, fmt.Errorf("min delta frames cannot be negative: %d", cfg.MinDeltaFrames)
	}

	if cfg.MaxDeltaFrames <= 0 {
		return nil, fmt.Errorf("max delta frames must be positive: %d", cfg.MaxDeltaFrames)
	}

	if cfg.MaxDeltaFrames < cfg.MinDeltaFrames {
		return nil, fmt.Errorf(
			"max delta frames (%d) cannot be less than min delta frames (%d)",
			cfg.MaxDeltaFrames,
			cfg.MinDeltaFrames,
		)
	}

	if cfg.Fanout <= 0 {
		return nil, fmt.Errorf("fanout must be positive: %d", cfg.Fanout)
	}

	if len(peaks) == 0 {
		return nil, nil
	}

	sorted := make([]Peak, len(peaks))
	copy(sorted, peaks)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Time < sorted[j].Time
	})

	var landmarks []Landmark

	for i, anchor := range sorted {
		matched := 0

		for j := i + 1; j < len(sorted) && matched < cfg.Fanout; j++ {
			target := sorted[j]
			delta := target.Time - anchor.Time

			if delta > cfg.MaxDeltaFrames {
				break
			}

			if delta < cfg.MinDeltaFrames {
				continue
			}

			landmarks = append(landmarks, Landmark{
				AnchorFrequency: anchor.Frequency,
				TargetFrequency: target.Frequency,
				DeltaTime:       delta,
				AnchorTime:      anchor.Time,
			})

			matched++
		}
	}

	return landmarks, nil
}
