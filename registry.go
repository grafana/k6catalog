package k6catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
)

var (
	ErrInvalidRegistry = fmt.Errorf("invalid module registry")     //nolint:revive
	ErrEntryNotFound   = fmt.Errorf("entry not found in registry") //nolint:revive
)

// Registry defines the interface of a module registry
type Registry interface {
	// GetVersion returns the versions of a a module given its name
	GetVersions(cxt context.Context, mod string) (Entry, error)
}

// Entry defines a registry entry
type Entry struct {
	Module   string   `json:"module,omitempty"`
	Versions []string `json:"versions,omitempty"`
}

type registry struct {
	dependencies map[string]Entry
}

// loadRegistryFromJSON returns a Registry from a json file
func loadRegistryFromJSON(path string) (Registry, error) {
	buff, err := os.ReadFile(path) //nolint:forbidigo,gosec
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRegistry, err)
	}

	dependencies := map[string]Entry{}
	err = json.Unmarshal(buff, &dependencies)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRegistry, err)
	}

	return registry{
		dependencies: dependencies,
	}, nil
}

// GetVersions returns the versions for a given module
func (r registry) GetVersions(_ context.Context, mod string) (Entry, error) {
	entry, found := r.dependencies[mod]
	if !found {
		return Entry{}, fmt.Errorf("%w : %s", ErrEntryNotFound, mod)
	}

	return entry, nil
}
