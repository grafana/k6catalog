// Package k6catalog defines the extension catalog service
package k6catalog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	"github.com/Masterminds/semver/v3"
)

const (
	defaultCatalogFile = "k6catalog.json"
)

var (
	ErrCannotSatisfy     = errors.New("cannot satisfy dependency") //nolint:revive
	ErrInvalidConstrain  = errors.New("invalid constrain")         //nolint:revive
	ErrUnknownDependency = errors.New("unknown dependency")        //nolint:revive
	ErrDownload          = errors.New("downloading catalog")       //nolint:revive
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

// NewCatalog creates a catalog from a registry
func NewCatalog(registry Registry) Catalog {
	return catalog{registry: registry}
}

// NewCatalogFromJSON creates a Catalog from a json file
func NewCatalogFromJSON(catalogFile string) (Catalog, error) {
	registry, err := loadRegistryFromJSON(catalogFile)
	if err != nil {
		return nil, err
	}
	return catalog{registry: registry}, nil
}

// NewCatalogFromURL creates a Catalog from a URL
func NewCatalogFromURL(ctx context.Context, catalogURL string) (Catalog, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, catalogURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w %s", ErrDownload, resp.Status)
	}

	catalogFile, err := os.CreateTemp("", "catalog*.json") //nolint:forbidigo
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}

	_, err = io.Copy(catalogFile, resp.Body)
	if err != nil {
		_ = catalogFile.Close()
		_ = os.Remove(catalogFile.Name()) //nolint:forbidigo
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}

	err = catalogFile.Close()
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}

	catalog, err := NewCatalogFromJSON(catalogFile.Name())
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}

	return catalog, nil
}

// DefaultCatalog creates a Catalog from the default json file 'catalog.json'
func DefaultCatalog() (Catalog, error) {
	return NewCatalogFromJSON(defaultCatalogFile)
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

	versions := []*semver.Version{}
	for _, v := range entry.Versions {
		version, err := semver.NewVersion(v)
		if err != nil {
			return Module{}, err
		}
		versions = append(versions, version)
	}

	if len(versions) > 0 {
		// try to find the higher version that satisfies the condition
		sort.Sort(sort.Reverse(semver.Collection(versions)))
		for _, v := range versions {
			if constrain.Check(v) {
				return Module{Path: entry.Module, Version: v.Original()}, nil
			}
		}
	}

	return Module{}, fmt.Errorf("%w : %s %s", ErrCannotSatisfy, dep.Name, dep.Constrains)
}
