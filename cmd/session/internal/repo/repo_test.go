package repo_test

import (
	"net"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/cmd/session/internal/app"
	"github.com/Meat-Hook/back-template/cmd/session/internal/repo"
	"github.com/Meat-Hook/back-template/libs/metrics"
	"github.com/gofrs/uuid"
)

func TestRepo_Smoke(t *testing.T) {
	t.Parallel()

	db, assert := start(t)

	m := metrics.DB("session", metrics.MethodsOf(&repo.Repo{})...)
	r := repo.New(db, &m)

	session := app.Session{
		ID: uuid.Must(uuid.NewV4()),
		Origin: app.Origin{
			IP:        net.ParseIP("192.100.10.4"),
			UserAgent: "Mozilla/5.0",
		},
		Token: app.Token{
			Value: "token",
		},
		UserID:    uuid.Must(uuid.NewV4()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := r.Save(ctx, session)
	assert.Nil(err)

	res, err := r.ByID(ctx, session.ID)
	assert.Nil(err)
	session.CreatedAt = res.CreatedAt
	session.UpdatedAt = res.UpdatedAt
	if session.Origin.IP.Equal(res.Origin.IP) {
		session.Origin.IP = res.Origin.IP
	}
	assert.Equal(session, *res)
	err = r.Delete(ctx, session.ID)
	assert.Nil(err)

	res, err = r.ByID(ctx, session.ID)
	assert.Nil(res)
	assert.ErrorIs(err, app.ErrNotFound)
}
