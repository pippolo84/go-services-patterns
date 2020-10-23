package model

import "sort"

// Analytics represents a simulated persistent memory for our analytics info
type Analytics struct {
	stats []*Score
}

// Score holds the score info about a single message
type Score struct {
	ThreadID  string `json:"thread_id"`
	MessageID string `json:"message_id"`
	Votes     int    `json:"votes"`
}

// NewAnalyticsStore returns an empty new Score store
func NewAnalyticsStore() *Analytics {
	return &Analytics{}
}

// Init initializes the store
func (a *Analytics) Init() error {
	return nil
}

// Close frees all resources used by te store
func (a *Analytics) Close() {}

// AddScore adds the new score information to the store
// it does not change anything if a record with the same MessageID and ThreadID
// and a greater score is already present
func (a *Analytics) AddScore(s *Score) {
	for i, score := range a.stats {
		if s.ThreadID == score.ThreadID && s.MessageID == score.MessageID {
			if s.Votes > score.Votes {
				a.stats[i] = s
			}
			return
		}
	}

	a.stats = append(a.stats, s)
}

// HighestScore returns a reference to the object with the highest score
// in the analytics store.
// It returns nil if the store is empty
func (a *Analytics) HighestScore() *Score {
	if len(a.stats) == 0 {
		return nil
	}

	sort.Slice(a.stats, func(i, j int) bool {
		return a.stats[i].Votes > a.stats[j].Votes
	})

	return a.stats[0]
}
