package k6catalog

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
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
			dep:    Dependency{Name: "dep", Constrains: "v0.1.0"},
			expect: Module{Path: "github.com/dep", Version: "v0.1.0"},
		},
		{
			title:  "resolve > constrain",
			dep:    Dependency{Name: "dep", Constrains: ">v0.1.0"},
			expect: Module{Path: "github.com/dep", Version: "v0.2.0"},
		},
		{
			title:  "resolve latest version",
			dep:    Dependency{Name: "dep", Constrains: "*"},
			expect: Module{Path: "github.com/dep", Version: "v0.2.0"},
		},
		{
			title:     "unsatisfied > constrain",
			dep:       Dependency{Name: "dep", Constrains: ">v0.2.0"},
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

const testCatalog = `{
	"k6/x/output-kafka": {"Module": "github.com/grafana/xk6-output-kafka", "Versions": ["v0.1.0", "v0.2.0"]}
}`

func TestCatalogFromURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		handler   http.HandlerFunc
		expectErr error
	}{
		{
			name: "download catalog",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(testCatalog))
			},
			expectErr: nil,
		},
		{
			name: "catalog not found",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectErr: ErrDownload,
		},
		{
			name: "empty catalog",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			expectErr: ErrInvalidRegistry,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tc.handler)

			_, err := NewCatalogFromURL(context.TODO(), srv.URL)

			if !errors.Is(err, tc.expectErr) {
				t.Fatalf("expected %v got %v", tc.expectErr, err)
			}
		})
	}
}
