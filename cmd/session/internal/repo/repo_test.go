package repo_test

import (
	"net"
	"testing"
	"time"

	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
	repo2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/repo"
	metrics2 "github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/gofrs/uuid"
	"github.com/rs/xid"
)

func TestRepo_Smoke(t *testing.T) {
	t.Parallel()

	db, assert := start(t)

	m := metrics2.DB("session", metrics2.MethodsOf(&repo2.Repo{})...)
	r := repo2.New(db, &m)

	session := app2.Session{
		ID: xid.New().String(),
		Origin: app2.Origin{
			IP:        net.ParseIP("192.100.10.4"),
			UserAgent: "Mozilla/5.0",
		},
		Token: app2.Token{
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
	assert.ErrorIs(err, app2.ErrNotFound)
}
