// Package cmd contains build cobra command factory function.
package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/grafana/k6catalog"
	"github.com/spf13/cobra"
)

var ErrTargetPlatformUndefined = errors.New("target platform is required") //nolint:revive

const long = `
Resolves dependencies considering version constraints
`

const example = `

k6catalog resolve -r registry.json -d k6/x/output-kafka -c >v0.7.0
github.com/grafana/xk6-output-kafka v0.8.0
`

// New creates new cobra command for resolve command.
func New() *cobra.Command {
	var (
		path       string
		dependency string
		constrains string
	)

	cmd := &cobra.Command{
		Use:     "resolve",
		Short:   "resolve dependencies",
		Long:    long,
		Example: example,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if path == "" {
				return fmt.Errorf("path to registry must be specified")
			}

			catalog, err := k6catalog.NewCatalog(context.TODO(), path)
			if err != nil {
				return err
			}

			result, err := catalog.Resolve(cmd.Context(), k6catalog.Dependency{Name: dependency, Constrains: constrains})
			if err != nil {
				return err
			}

			fmt.Printf("%s %s\n", result.Path, result.Version)

			return nil
		},
	}

	cmd.Flags().StringVar(
		&path,
		"catalog",
		k6catalog.DefaultCatalogURL,
		"path to catalog. Can be a path to a local file or a URL",
	)
	cmd.Flags().StringVarP(&dependency, "name", "d", "", "name of dependency")
	cmd.Flags().StringVarP(&constrains, "constrains", "c", "*", "version constrains")

	return cmd
}
