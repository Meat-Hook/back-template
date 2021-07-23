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
	reg = prometheus.NewPedanticRegistry()
)

func TestMain(m *testing.M) {
	metrics.InitMetrics(reg)

	os.Exit(m.Run())
}

func start(t *testing.T) (context.Context, *repo.Repo, *require.Assertions) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)

	r := require.New(t)
	pool, err := dockertest.NewPool("")
	r.Nil(err)
	pool.MaxWait = timeout

	opt := &dockertest.RunOptions{
		Repository: "cockroachdb/cockroach",
		Tag:        "v20.2.0",
		Cmd:        []string{"start-single-node", "--insecure"},
	}

	resource, err := pool.RunWithOptions(opt, func(cfg *docker.HostConfig) {
		cfg.AutoRemove = true
	})
	r.Nil(err)

	m := db.NewMetrics(reg, "test", "repo", new(repo.Repo))
	var conn *db.DB
	err = pool.Retry(func() error {
		str := fmt.Sprintf("host=localhost port=%s user=root "+
			"password=root dbname=postgres sslmode=disable", resource.GetPort("26257/tcp"))
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
	r.Nil(err)

	t.Cleanup(func() {
		err = pool.Purge(resource)
		r.Nil(err)
	})

	return ctx, repo.New(conn), r
}
