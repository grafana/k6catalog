// Package cmd contains build cobra command factory function.
package cmd

import (
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

k6catalog -r registry.json -d k6/x/output-kafka -c >v0.7.0
github.com/grafana/xk6-output-kafka v0.8.0
`

// New creates new cobra command for build command.
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

			registry, err := k6catalog.NewJSONRegistry(path)
			if err != nil {
				return err
			}

			catalog := k6catalog.NewCatalog(registry)

			result, err := catalog.Resolve(cmd.Context(), k6catalog.Dependency{Module: dependency, Constrains: constrains})
			if err != nil {
				return err
			}

			fmt.Printf("%s %s\n", result.Path, result.Version)

			return nil
		},
	}

	cmd.Flags().StringVarP(&path, "registry", "r", "", "path to registry")
	cmd.Flags().StringVarP(&dependency, "name", "d", "", "name of dependency")
	cmd.Flags().StringVarP(&constrains, "constrains", "c", "*", "version constrains")

	return cmd
}
