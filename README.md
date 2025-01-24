> Deprecation notice: the catalog functionality [has been merged](https://github.com/grafana/k6build/pull/91) into the https://github.com/grafana/k6build repository is no longer used.


# k6catalog

k6catalog is a library that maps a dependency for k6 extension with optional semantic versioning constrains to the corresponding golang modules.

For example `k6/x/output-kafka:>0.1.0` ==> `github.com/grafana/xk6-output-kafka@v0.2.0`

See documentation https://pkg.go.dev/github.com/grafana/k6catalog

## JSON file catalog

The `NewCatalogFromJSON` function creates a catalog from a JSON file that has the following [schema](./schema.json):

```
{
        "<dependency>": {"module": "<go module path>", "versions": ["<version>", "<version>", ... "<version>"]}
}
```
where:
- `<dependency>` is the import path for the dependency
- `module` is the path to the go module that implements the dependency
- `versions` is the list of supported versions

Example:

```json
{
        "k6": {"module": "go.k6.io/k6", "versions": ["v0.50.0", "v0.51.0"]},
        "k6/x/kubernetes": {"module": "github.com/grafana/xk6-kubernetes", "versions": ["v0.8.0","v0.9.0"]},
        "k6/x/output-kafka": {"module": "github.com/grafana/xk6-output-kafka", "versions": ["v0.7.0"]}
}
```
 
