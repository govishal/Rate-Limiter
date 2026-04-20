package models

import "time"

// UserStats is per-user rate limiter counters exposed via GET /stats.
type UserStats struct {
	UserID           string    `json:"user_id"`
	AcceptedRequests int       `json:"accepted_requests"`
	RejectedRequests int       `json:"rejected_requests"`
	LastRequestAt    time.Time `json:"last_request_at,omitempty"`
	CurrentWindow    int       `json:"current_window_request_count"`
}
