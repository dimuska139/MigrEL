# MigrEL

## About
An Elasticsearch Migration Tool.
A utility that manages structural changes to your Elasticsearch.
It applies migrations in correct order to indexes.

## Configuration
The application by default looks for the configuration file in the same directory
where its binary file is located. But you can specify an arbitrary
path using the `--config` parameter.
### Configuration file structure
```yaml
elasticsearch:
  # Index in which information about applied migrations will be stored
  # (default: migrations)
  migrations_index_name: migrations  # Optional
  host: http://127.0.0.1:9217
  username:  # Optional
  password:  # Optional
  tls:  # Optional
    cert_file:
    key_file:
```

## How to use
1. Install the goose binary to your `$GOPATH/bin` directory: `go install github.com/pressly/goose/v3/cmd/goose@latest`
2. Create a file with migration: `goose -dir ./migrations create create_index_accounts`.
`create_index_accounts` - arbitrary name of the migration
3. Implement migration in a created file `./migrations/XXX_create_index_accounts.go`
4. Create a configuration file `cp config.yml.dist config.yml`
5. Fill the configuration file (`config.yml`) with the necessary values
6. Run application with `up` (apply migrations) or `down` (rollback migrations) option: `go run ./cmd/migrel/main.go up`
7. Check stdout and make sure the changes are applied in your database

You can run Elasticsearch in Docker to check your migrations locally.
See [docker-compose.yml](docker-compose.yml).

## Warning
* Define your own types, structures and constants only
in your up-migration functions (to avoid collisions)
* You should use global variable `elasticsearch.Client`
from package `github.com/dimuska139/migrel/pkg/elasticsearch` in your migrations to communicate
with the Elasticsearch instance
* You shouldn't use `tx *sql.Tx` parameter in your migrations

## Migration example
```go
package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/pressly/goose/v3"

	"github.com/dimuska139/migrel/pkg/elasticsearch"
)

func init() {
	goose.AddMigrationContext(upCreateIndexAccounts, downCreateIndexAccounts)
}

func upCreateIndexAccounts(ctx context.Context, _ *sql.Tx) error {
	_, err := elasticsearch.Client.Indices.Create("accounts").
		Request(&create.Request{
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"name":       types.NewTextProperty(),
					"email":      types.NewTextProperty(),
					"created_at": types.NewDateProperty(),
				},
			},
		}).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	return nil
}

func downCreateIndexAccounts(ctx context.Context, _ *sql.Tx) error {
	_, err := elasticsearch.Client.Indices.Delete("accounts").
		Do(ctx)
	if err != nil {
		return fmt.Errorf("delete index: %w", err)
	}
	
	return nil
}

```