package storage

import (
	"errors"
	"sync"
)

// TargetProvider provides thread-safe access to the active default Location.
type TargetProvider interface {
	Current() (Location, error)
	Set(Location) error
}

// MutableTargetProvider is a thread-safe implementation of TargetProvider.
type MutableTargetProvider struct {
	mu       sync.RWMutex
	location Location
	hasValue bool
}

// NewMutableTargetProvider creates a new MutableTargetProvider initialized with initial Location.
func NewMutableTargetProvider(initial Location) (*MutableTargetProvider, error) {
	if err := initial.Validate(); err != nil {
		return nil, err
	}

	return &MutableTargetProvider{
		location: initial,
		hasValue: true,
	}, nil
}

func (p *MutableTargetProvider) Current() (Location, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.hasValue {
		return Location{}, errors.New("download location has not been configured")
	}

	return p.location, nil
}

func (p *MutableTargetProvider) Set(location Location) error {
	if err := location.Validate(); err != nil {
		return err
	}

	p.mu.Lock()
	p.location = location
	p.hasValue = true
	p.mu.Unlock()

	return nil
}
