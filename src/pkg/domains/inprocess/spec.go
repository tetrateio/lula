// Copyright (c) Tetrate, Inc 2025 All Rights Reserved.

package inprocess

import (
	"context"

	"github.com/defenseunicorns/lula/src/types"
)

type (
	// Spec configures the properties of the in process domain
	Spec struct {
		// Resources is a map of variables to function names to call
		Resources map[string]string `json:"resources" yaml:"resources"`
	}

	// Domain is a structure that contains the domain type and the corresponding spec
	Domain struct {
		Spec     *Spec `json:"spec" yaml:"spec"`
		registry Registry
	}
)

// CreateDomain creates a new in process domain.
func CreateDomain(spec *Spec) (*Domain, error) {
	return &Domain{
		Spec:     spec,
		registry: GlobalRegistry(),
	}, nil
}

// IsExecutable implements Domain.
func (d Domain) IsExecutable() bool { return false }

// GetResources implements Domain.
func (d Domain) GetResources(ctx context.Context) (types.DomainResources, error) {
	resources := make(types.DomainResources)
	for variable, functionName := range d.Spec.Resources {
		res, err := d.registry.Run(ctx, functionName)
		if err != nil {
			return nil, err
		}
		resources[variable] = res
	}
	return resources, nil
}

var _ types.Domain = (*Domain)(nil)
