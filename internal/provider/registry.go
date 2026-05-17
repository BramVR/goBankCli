package provider

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrInvalidName    = errors.New("provider name is required")
	ErrAlreadyExists  = errors.New("provider already registered")
	ErrProviderAbsent = errors.New("provider not registered")
)

type Registry struct {
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: map[string]Factory{}}
}

func (r *Registry) Register(name string, factory Factory) error {
	key := normalizeName(name)
	if key == "" {
		return ErrInvalidName
	}
	if factory == nil {
		return errors.New("provider factory is required")
	}
	if _, ok := r.factories[key]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyExists, key)
	}
	r.factories[key] = factory
	return nil
}

func (r *Registry) Create(name string, cfg Config) (Provider, error) {
	key := normalizeName(name)
	factory, ok := r.factories[key]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderAbsent, key)
	}
	return factory(cfg)
}

func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.factories))
	for name := range r.factories {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func normalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
