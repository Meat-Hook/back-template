package repo_test

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/repo"
	"github.com/rs/xid"
)

func TestRepo_Smoke(t *testing.T) {
	t.Parallel()

	db, assert := start(t)

	m := metrics.DB("session", metrics.MethodsOf(&repo.Repo{})...)
	r := repo.New(db, &m)

	session := app.Session{
		ID: xid.New().String(),
		Origin: app.Origin{
			IP:        net.ParseIP("192.100.10.4"),
			UserAgent: "Mozilla/5.0",
		},
		Token: app.Token{
			Value: "token",
		},
		UserID:    1,
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
	assert.Equal(app.ErrNotFound, errors.Unwrap(err))
}
