package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/pressly/goose/v3/database"
	"strconv"
	"time"
)

type Migration struct {
	Version   int64     `json:"version"`
	IsApplied bool      `json:"is_applied"`
	CreatedAt time.Time `json:"created_at"`
}

func NewStore(tableName string, esClient *elasticsearch.TypedClient) *Store {
	return &Store{
		tableName: tableName,
		esClient:  esClient,
	}
}

type Store struct {
	tableName string
	esClient  *elasticsearch.TypedClient
}

func (s *Store) Tablename() string {
	return s.tableName
}

// CreateVersionTable creates the version table, which is used to track migrations.
func (s *Store) CreateVersionTable(ctx context.Context, _ database.DBTxConn) error {
	_, err := s.esClient.Indices.Create(s.tableName).
		Request(&create.Request{
			Mappings: &types.TypeMapping{
				Properties: map[string]types.Property{
					"version":    types.NewLongNumberProperty(),
					"created_at": types.NewDateProperty(),
					"is_applied": types.NewBooleanProperty(),
				},
			},
		}).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("create index: %w", err)
	}

	return nil
}

// Insert a version id into the version table.
func (s *Store) Insert(ctx context.Context, _ database.DBTxConn, req database.InsertRequest) error {
	_, err := s.esClient.Index(s.tableName).
		Request(Migration{
			Version:   req.Version,
			CreatedAt: time.Now(),
			IsApplied: true,
		}).
		Refresh(refresh.True).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("index: %w", err)
	}

	return nil
}

// Delete removes a version id from the version table.
func (s *Store) Delete(ctx context.Context, _ database.DBTxConn, version int64) error {
	res, err := s.esClient.Search().
		Index(s.tableName).
		Request(&search.Request{
			Query: &types.Query{
				Match: map[string]types.MatchQuery{
					"version_id": {
						Query: strconv.Itoa(int(version)),
					},
				},
			},
		}).Do(ctx)
	if err != nil {
		return fmt.Errorf("search: %w", err)
	}

	var documentID string
	for _, hit := range res.Hits.Hits {
		documentID = hit.Id_
		break
	}

	if documentID == "" {
		return nil
	}

	_, err = s.esClient.Delete(s.tableName, documentID).Do(ctx)

	return nil
}

// GetMigration retrieves a single migration by version id. If the query succeeds, but the
// version is not found, this method must return [ErrVersionNotFound].
func (s *Store) GetMigration(
	ctx context.Context,
	_ database.DBTxConn,
	version int64,
) (*database.GetMigrationResult, error) {
	res, err := s.esClient.Search().
		Index(s.tableName).
		Request(&search.Request{
			Query: &types.Query{
				Match: map[string]types.MatchQuery{
					"version": {Query: strconv.Itoa(int(version))},
				},
			},
		}).Do(ctx)
	if err != nil {
		var elErr *types.ElasticsearchError
		if errors.As(err, &elErr) && elErr.Status == 404 {
			return nil, fmt.Errorf("%w: %d", database.ErrVersionNotFound, version)
		}
		return nil, fmt.Errorf("get migration: %w", err)
	}

	var document Migration
	for _, hit := range res.Hits.Hits {
		if err := json.Unmarshal(hit.Source_, &document); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		break
	}

	if document == (Migration{}) {
		return nil, database.ErrVersionNotFound
	}

	return &database.GetMigrationResult{
		Timestamp: document.CreatedAt,
		IsApplied: document.IsApplied,
	}, nil
}

// GetLatestVersion retrieves the last applied migration version. If no migrations exist, this
// method must return [ErrVersionNotFound].
func (s *Store) GetLatestVersion(ctx context.Context, _ database.DBTxConn) (int64, error) {
	res, err := s.esClient.Search().
		Index(s.tableName).
		Request(&search.Request{
			Sort: []types.SortCombinations{
				"created_at:desc",
			},
		}).Do(ctx)
	if err != nil {
		var elErr *types.ElasticsearchError
		if errors.As(err, &elErr) && elErr.Status == 404 {
			return 0, database.ErrVersionNotFound
		}
		return 0, fmt.Errorf("search: %w", err)
	}

	var document Migration
	for _, hit := range res.Hits.Hits {
		if err := json.Unmarshal(hit.Source_, &document); err != nil {
			return 0, fmt.Errorf("unmarshal: %w", err)
		}
		break
	}

	if document == (Migration{}) {
		return 0, database.ErrVersionNotFound
	}

	return document.Version, nil
}

// ListMigrations retrieves all migrations sorted in descending order by id or timestamp. If
// there are no migrations, return empty slice with no error. Typically this method will return
// at least one migration, because the initial version (0) is always inserted into the version
// table when it is created.
func (s *Store) ListMigrations(ctx context.Context, _ database.DBTxConn) ([]*database.ListMigrationsResult, error) {
	res, err := s.esClient.Search().
		Index(s.tableName).
		Request(&search.Request{
			Sort: []types.SortCombinations{
				types.SortOptions{SortOptions: map[string]types.FieldSort{
					"created_at": {
						Order: &sortorder.Desc},
				}},
			},
		}).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	var results []*database.ListMigrationsResult
	for _, hit := range res.Hits.Hits {
		var document Migration
		if err := json.Unmarshal(hit.Source_, &document); err != nil {
			return nil, fmt.Errorf("unmarshal: %w", err)
		}
		results = append(results, &database.ListMigrationsResult{
			Version:   document.Version,
			IsApplied: document.IsApplied,
		})
	}

	return results, nil
}
