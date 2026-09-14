package matcher

import (
	"fmt"
	"sort"

	"github.com/sidx04/bassclef/internal/storage"
)

// Match ranks catalog songs against a query clip's landmarks using
// time-offset voting: for each matching hash, the offset
// between the catalog's and the query's time is tallied per song, and a
// song's score is its single best-aligned offset's vote count.
func Match(store storage.Store, query []QueryLandmark, cfg MatchConfig) ([]Candidate, error) {
	if cfg.TopN <= 0 {
		return nil, fmt.Errorf("top N must be positive: %d", cfg.TopN)
	}

	if cfg.MinVotes < 0 {
		return nil, fmt.Errorf("min votes cannot be negative: %d", cfg.MinVotes)
	}

	if len(query) == 0 {
		return nil, nil
	}

	votes := make(map[int64]map[int]int)

	for _, q := range query {
		matches, err := store.LookupHash(q.Hash)
		if err != nil {
			return nil, fmt.Errorf("failed to look up hash: %w", err)
		}

		for _, m := range matches {
			offset := m.TimeOffset - q.Time

			if votes[m.SongID] == nil {
				votes[m.SongID] = make(map[int]int)
			}
			votes[m.SongID][offset]++
		}
	}

	var candidates []Candidate

	for songID, offsets := range votes {
		best := 0
		for _, count := range offsets {
			if count > best {
				best = count
			}
		}

		if best < cfg.MinVotes {
			continue
		}

		candidates = append(candidates, Candidate{SongID: songID, Score: best})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score != candidates[j].Score {
			return candidates[i].Score > candidates[j].Score
		}
		return candidates[i].SongID < candidates[j].SongID
	})

	if len(candidates) > cfg.TopN {
		candidates = candidates[:cfg.TopN]
	}

	return candidates, nil
}
