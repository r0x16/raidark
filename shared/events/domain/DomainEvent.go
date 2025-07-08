package domain

import (
	"context"
	"time"
)

type DomainEvent interface {
	Name() string
	OccurredAt() time.Time
}

type EventHandler func(ctx context.Context, evt DomainEvent) error
