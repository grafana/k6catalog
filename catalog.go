// Package k6catalog defines the extension catalog service
package k6catalog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
)

const (
	DefaultCatalogFile = "catalog.json"                        //nolint:revive
	DefaultCatalogURL  = "https://registry.k6.io/catalog.json" //nolint:revive
)

var (
	ErrCannotSatisfy     = errors.New("cannot satisfy dependency") //nolint:revive
	ErrDownload          = errors.New("downloading catalog")       //nolint:revive
	ErrInvalidConstrain  = errors.New("invalid constrain")         //nolint:revive
	ErrInvalidCatalog    = fmt.Errorf("invalid catalog")           //nolint:revive
	ErrOpening           = errors.New("opening catalog")           //nolint:revive
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

// entry defines a catalog entry
type entry struct {
	Module   string   `json:"module,omitempty"`
	Versions []string `json:"versions,omitempty"`
}

type catalog struct {
	dependencies map[string]entry
}

// getVersions returns the versions for a given module
func (c catalog) getVersions(_ context.Context, mod string) (entry, error) {
	e, found := c.dependencies[mod]
	if !found {
		return entry{}, fmt.Errorf("%w : %s", ErrUnknownDependency, mod)
	}

	return e, nil
}

// NewCatalogFromJSON creates a Catalog from a json file
func NewCatalogFromJSON(stream io.Reader) (Catalog, error) {
	buff := &bytes.Buffer{}
	_, err := buff.ReadFrom(stream)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCatalog, err)
	}

	dependencies := map[string]entry{}
	err = json.Unmarshal(buff.Bytes(), &dependencies)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidCatalog, err)
	}

	return catalog{
		dependencies: dependencies,
	}, nil
}

// NewCatalog returns a catalog loaded from a location.
// The location can be a local path or an URL
func NewCatalog(ctx context.Context, location string) (Catalog, error) {
	if strings.HasPrefix(location, "http") {
		return NewCatalogFromURL(ctx, location)
	}

	return NewCatalogFromFile(location)
}

// NewCatalogFromFile creates a Catalog from a json file
func NewCatalogFromFile(catalogFile string) (Catalog, error) {
	json, err := os.ReadFile(catalogFile) //nolint:forbidigo,gosec
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrOpening, err)
	}

	buff := bytes.NewBuffer(json)
	return NewCatalogFromJSON(buff)
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

	catalog, err := NewCatalogFromJSON(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %w", ErrDownload, err)
	}

	return catalog, nil
}

// DefaultCatalog creates a Catalog from the default catalog URL
func DefaultCatalog() (Catalog, error) {
	return NewCatalogFromURL(context.TODO(), DefaultCatalogURL)
}

func (c catalog) Resolve(ctx context.Context, dep Dependency) (Module, error) {
	entry, err := c.getVersions(ctx, dep.Name)
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
