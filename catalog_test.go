package k6catalog

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type fakeRegistry struct {
	entries map[string]Entry
}

func (r fakeRegistry) GetVersions(_ context.Context, dep string) (Entry, error) {
	e, found := r.entries[dep]
	if !found {
		return Entry{}, fmt.Errorf("entry not found : %s", dep)
	}

	return Entry{Module: e.Module, Versions: e.Versions}, nil
}

func TestResolve(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		title     string
		dep       Dependency
		expect    Module
		expectErr error
	}{
		{
			title:  "resolve exact version",
			dep:    Dependency{Module: "dep", Constrains: "v0.1.0"},
			expect: Module{Path: "github.com/dep", Version: "v0.1.0"},
		},
		{
			title:  "resolve > constrain",
			dep:    Dependency{Module: "dep", Constrains: ">v0.1.0"},
			expect: Module{Path: "github.com/dep", Version: "v0.2.0"},
		},
		{
			title:     "unsatisfied > constrain",
			dep:       Dependency{Module: "dep", Constrains: ">v0.2.0"},
			expectErr: ErrCannotSatisfy,
		},
	}

	registry := fakeRegistry{
		entries: map[string]Entry{
			"dep": {Module: "github.com/dep", Versions: []string{"v0.1.0", "v0.2.0"}},
		},
	}

	catalog := NewCatalog(registry)
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.title, func(t *testing.T) {
			t.Parallel()

			mod, err := catalog.Resolve(context.TODO(), tc.dep)
			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}

			if tc.expectErr == nil && mod != tc.expect {
				t.Fatalf("expected %v got %v", tc.expect, mod)
			}
		})
	}
}
