package domain

import "github.com/r0x16/Raidark/shared/events/domain"

type ApiModule interface {
	Name() string
	Setup() error
	GetModel() []any
	GetSeedData() []any
	GetEventListeners() []domain.EventListener
}
