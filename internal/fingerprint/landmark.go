package fingerprint

import (
	"cmp"
	"fmt"
	"slices"
)

// GenerateLandmarks pairs each peak, as an anchor, with the nearest-in-time
// peaks that follow it within cfg's delta window, up to cfg.Fanout pairs
// per anchor. peaks need not be time-sorted: if they are not, a sorted
// copy is made internally. peaks itself is never modified.
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

	sortByTimeFunc := func(a, b Peak) int {
		return cmp.Compare(a.Time, b.Time)
	}

	sortedPeaks := peaks

	// FindPeaks already returns peaks in time order, so the copy and sort
	// are usually skipped. this also reduces the memory overhead
	// of creating duplicated slices everytime GenerateLandmarks is called.
	if !slices.IsSortedFunc(peaks, sortByTimeFunc) {
		sortedPeaks = slices.Clone(peaks)
		slices.SortStableFunc(sortedPeaks, sortByTimeFunc)
	}

	var landmarks = make([]Landmark, 0, len(peaks)*cfg.Fanout)

	for i, anchor := range sortedPeaks {
		matched := 0

		for j := i + 1; j < len(sortedPeaks) && matched < cfg.Fanout; j++ {
			target := sortedPeaks[j]
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
