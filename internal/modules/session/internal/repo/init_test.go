package repo_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Meat-Hook/migrate/core"
	"github.com/Meat-Hook/migrate/fs"
	"github.com/Meat-Hook/migrate/migrater"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

const (
	migrateDir = `../../migrate`
	timeout    = time.Second * 5
)

var (
	ctx context.Context
)

func start(t *testing.T) (*sqlx.DB, *require.Assertions) {
	r := require.New(t)
	pool, err := dockertest.NewPool("")
	r.Nil(err)

	resource, err := pool.Run("postgres", "latest", []string{
		"POSTGRES_USER=postgres",
		"POSTGRES_DB=postgres",
		"POSTGRES_PASSWORD=postgres",
	})
	r.Nil(err)

	var db *sqlx.DB
	if err := pool.Retry(func() error {
		str := fmt.Sprintf("host=localhost port=%s user=postgres "+
			"password=postgres dbname=postgres sslmode=disable", resource.GetPort("5432/tcp"))
		db, err = sqlx.Connect("postgres", str)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		r.Nil(err)
	}

	t.Cleanup(func() {
		err = pool.Purge(resource)
		r.Nil(err)
	})

	migrate := core.New(fs.New(), migrater.New(db.DB, logrus.New()))

	var cancel func()
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	err = migrate.Migrate(ctx, migrateDir, core.Config{Cmd: core.Up})
	r.Nil(err)

	return db, r
}
