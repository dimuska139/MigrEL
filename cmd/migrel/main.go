package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/pressly/goose/v3"
	"github.com/urfave/cli"
	"os"

	_ "modernc.org/sqlite"

	"github.com/dimuska139/migrel/internal/config"
	"github.com/dimuska139/migrel/internal/store"
	"github.com/dimuska139/migrel/migrations"
	"github.com/dimuska139/migrel/pkg/elasticsearch"
	"github.com/dimuska139/migrel/pkg/logging"
)

func migrate(c *cli.Context) error {
	ctx, cancelCtx := context.WithCancel(context.Background())
	defer cancelCtx()

	cfg, err := config.NewConfig(c.String("config"))
	if err != nil {
		return fmt.Errorf("new config: %w", err)
	}
	logger := logging.NewLogger(logging.LogLevelInfo)

	elasticClient, err := elasticsearch.NewElasticsearchClient(elasticsearch.Config{
		Addresses: []string{cfg.Elasticsearch.Host},
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
		TLS: struct {
			CertFile string
			KeyFile  string
		}{CertFile: cfg.Elasticsearch.TLS.CertFile, KeyFile: cfg.Elasticsearch.TLS.KeyFile},
	})
	if err != nil {
		return fmt.Errorf("initialize Elasticsearch client: %w", err)
	}

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	goose.SetLogger(logger)

	p, err := goose.NewProvider(
		"",
		db,
		migrations.Embed,
		goose.WithStore(store.NewStore(cfg.Elasticsearch.MigrationsIndexName, elasticClient)),
	)
	if err != nil {
		return fmt.Errorf("new provider: %w", err)
	}

	if c.Command.Name == "up" {
		if _, err := p.Up(ctx); err != nil {
			return fmt.Errorf("up: %w", err)
		}
	} else if c.Command.Name == "down" {
		if _, err := p.Down(ctx); err != nil {
			return fmt.Errorf("down: %w", err)
		}
	}

	version, err := p.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("get db version: %w", err)
	}

	logger.Info(fmt.Sprintf("Current database version is: %d", version))

	return nil
}

func main() {
	app := &cli.App{
		Name:  "Elasticsearch migrator",
		Usage: "Migrator for the Elasticsearch",
		Commands: []cli.Command{
			{
				Name:  "up",
				Usage: "Apply migrations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Value: "./config.yml",
						Usage: "path to the config file",
					},
				},
				Action: migrate,
			},
			{
				Name:  "down",
				Usage: "Rollback migrations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "config",
						Value: "./config.yml",
						Usage: "path to the config file",
					},
				},
				Action: migrate,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger := logging.NewLogger(logging.LogLevelInfo)
		logger.Fatal("Can't migrate: ", err.Error())
	}
}
