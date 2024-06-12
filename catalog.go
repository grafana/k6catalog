// Package k6catalog defines the extension catalog service
package k6catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/semver"
)

var (
	ErrCannotSatisfy     = errors.New("cannot satisfy dependency") //nolint:revive
	ErrInvalidConstrain  = errors.New("invalid constrain")         //nolint:revive
	ErrUnknownDependency = errors.New("unknown dependency")        //nolint:revive
)

// Dependency defines a Dependency with a version constrain
// Examples:
// Name: k6/x/k6-kubernetes   Constrains *
// Name: k6/x/k6-output-kafka Constrains >v0.9.0
type Dependency struct {
	Name       string `json:"name,omitempty"`
	Constrains string `json:"constrains,omitempty"`
}

// Module defines a go module that resolves a Dependency
type Module struct {
	Path    string `json:"path,omitempty"`
	Version string `json:"version,omitempty"`
}

// Catalog defines the interface of the extension catalog service
type Catalog interface {
	// Resolve returns a Module that satisfies a Dependency
	Resolve(ctx context.Context, dep Dependency) (Module, error)
}

type catalog struct {
	registry Registry
}

// NewCatalog creates a Catalog from a registry
func NewCatalog(registry Registry) Catalog {
	return catalog{registry: registry}
}

func (c catalog) Resolve(ctx context.Context, dep Dependency) (Module, error) {
	entry, err := c.registry.GetVersions(ctx, dep.Name)
	if err != nil {
		return Module{}, err
	}

	constrain, err := semver.NewConstraint(dep.Constrains)
	if err != nil {
		return Module{}, fmt.Errorf("%w : %s", ErrInvalidConstrain, dep.Constrains)
	}

	for _, v := range entry.Versions {
		version, err := semver.NewVersion(v)
		if err != nil {
			return Module{}, err
		}

		if constrain.Check(version) {
			return Module{Path: entry.Module, Version: version.Original()}, nil
		}
	}

	return Module{}, fmt.Errorf("%w : %s %s", ErrCannotSatisfy, dep.Name, dep.Constrains)
}
