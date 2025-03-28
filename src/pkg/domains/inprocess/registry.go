// Copyright (c) Tetrate, Inc 2025 All Rights Reserved.

package inprocess

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrAlreadyRegistered = errors.New("already registered")
	ErrNotFound          = errors.New("not found")

	// globalRegistry is the global registry of domain functions.
	globalRegistry = NewRegistry()

	_ Registry = (*registry)(nil)
)

// ResourcesFunction is a function used to return the domain resources.
type (
	ResourcesFunction func(context.Context) (any, error)

	// Registry is the interface for the global registry of domain functions.
	Registry interface {
		// Register the given domain function in the registry.
		Register(name string, f ResourcesFunction) error
		// Unregister the given domain function from the registry.
		Unregister(name string)
		// Run the domain function with the given name.
		Run(ctx context.Context, name string) (any, error)
	}
)

// GlobalRegistry returns the global registry of domain functions.
func GlobalRegistry() Registry { return globalRegistry }

// NewRegistry creates a new registry of domain functions.
func NewRegistry() Registry {
	return &registry{functions: make(map[string]ResourcesFunction)}
}

// registry is a thread-safe collection of domain functions.
type registry struct {
	mu        sync.RWMutex
	functions map[string]ResourcesFunction
}

// Register the given domain function in the registry.
func (r *registry) Register(name string, f ResourcesFunction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.functions[name]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyRegistered, name)
	}
	r.functions[name] = f
	return nil
}

// Unregister the given domain function from the registry.
func (r *registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.functions, name)
}

// Run the domain function with the given name.
func (r *registry) Run(ctx context.Context, name string) (any, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	f, ok := r.functions[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return f(ctx)
}
