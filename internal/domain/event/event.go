package event

import "time"

type Event interface {
	Type() string
	OccurredAt() time.Time
}

type Base struct {
	EventType  string    `json:"type"`
	OccurredOn time.Time `json:"occurred_at"`
}

func (b Base) Type() string          { return b.EventType }
func (b Base) OccurredAt() time.Time { return b.OccurredOn }
