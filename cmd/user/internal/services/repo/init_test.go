package repo_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/Meat-Hook/back-template/cmd/user/internal/services/repo"
	"github.com/Meat-Hook/back-template/libs/db"
	"github.com/Meat-Hook/back-template/libs/metrics"
)

const (
	migrateDir = `../../../migrate`
	timeout    = time.Second * 30
)

var (
	logger = zerolog.New(os.Stdout)
	reg    = prometheus.NewPedanticRegistry()
)

func TestMain(m *testing.M) {
	metrics.InitMetrics(reg)

	os.Exit(m.Run())
}

func start(t *testing.T) (context.Context, *repo.Repo, *require.Assertions) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	assert := require.New(t)
	pool, err := dockertest.NewPool("")
	assert.NoError(err)
	pool.MaxWait = timeout

	opt := &dockertest.RunOptions{
		Repository: "cockroachdb/cockroach",
		Tag:        "v20.2.0",
		Cmd:        []string{"start-single-node", "--insecure"},
	}

	resource, err := pool.RunWithOptions(opt, func(cfg *docker.HostConfig) {
		cfg.AutoRemove = true
	})
	assert.NoError(err)

	m := db.NewMetrics(reg, "test", new(repo.Repo))
	var conn *db.DB
	err = pool.Retry(func() error {
		str := fmt.Sprintf("postgresql://root:root@localhost:%s/defaultdb?sslmode=disable", resource.GetPort("26257/tcp"))
		conn, err = db.Postgres(ctx, db.PostgresConfig{
			DSN:        str,
			MigrateDir: migrateDir,
			Metric:     *m,
		})
		if err != nil {
			return err
		}

		return nil
	})
	assert.NoError(err)

	t.Cleanup(func() {
		err = pool.Purge(resource)
		assert.NoError(err)
	})

	return logger.WithContext(ctx), repo.New(conn), assert
}
