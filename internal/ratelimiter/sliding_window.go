package ratelimiter

import (
	"sort"
	"sync"
	"time"

	"rate-limiter/internal/models"
)

// SlidingWindow enforces a max request count per user over a rolling time window
// using in-memory timestamps (single-process; mutex-safe).
type SlidingWindow struct {
	maxRequests int
	window      time.Duration

	mu      sync.Mutex
	buckets map[string][]time.Time
	stats   map[string]models.UserStats
}

// New returns an in-memory sliding-window rate limiter (mutex-safe, single-process).
func New(maxRequests int, window time.Duration) *SlidingWindow {
	return &SlidingWindow{
		maxRequests: maxRequests,
		window:      window,
		buckets:     make(map[string][]time.Time),
		stats:       make(map[string]models.UserStats),
	}
}

func (s *SlidingWindow) TryConsume(userID string) (bool, int) {
	now := time.Now().UTC()
	cutoff := now.Add(-s.window)

	s.mu.Lock()
	defer s.mu.Unlock()

	timestamps := timestampsAfterCutoff(s.buckets[userID], cutoff)

	stats := s.stats[userID]
	stats.LastRequestAt = now

	if len(timestamps) >= s.maxRequests {
		stats.RejectedRequests++
		stats.CurrentWindow = len(timestamps)
		s.stats[userID] = stats
		s.buckets[userID] = timestamps
		return false, len(timestamps)
	}

	timestamps = append(timestamps, now)
	stats.AcceptedRequests++
	stats.CurrentWindow = len(timestamps)
	s.stats[userID] = stats
	s.buckets[userID] = timestamps

	return true, len(timestamps)
}

// SnapshotStats returns one entry per known user, sorted by user_id for stable JSON.
func (s *SlidingWindow) SnapshotStats() []models.UserStats {
	now := time.Now().UTC()
	cutoff := now.Add(-s.window)

	s.mu.Lock()
	defer s.mu.Unlock()

	ids := make([]string, 0, len(s.stats))
	for userID := range s.stats {
		ids = append(ids, userID)
	}
	sort.Strings(ids)

	out := make([]models.UserStats, 0, len(ids))
	for _, userID := range ids {
		stats := s.stats[userID]
		inWindow := timestampsAfterCutoff(s.buckets[userID], cutoff)
		s.buckets[userID] = inWindow
		stats.CurrentWindow = len(inWindow)
		s.stats[userID] = stats
		row := stats
		row.UserID = userID
		out = append(out, row)
	}

	return out
}

func (s *SlidingWindow) MaxRequests() int {
	return s.maxRequests
}

func (s *SlidingWindow) Window() time.Duration {
	return s.window
}

// timestampsAfterCutoff keeps entries strictly after cutoff (re-slices in place when possible).
func timestampsAfterCutoff(timestamps []time.Time, cutoff time.Time) []time.Time {
	valid := timestamps[:0]
	for _, ts := range timestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	return valid
}
