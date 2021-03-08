package repo_test

import (
	"testing"
	"time"

	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/Meat-Hook/back-template/internal/microservices/user/internal/app"
	"github.com/Meat-Hook/back-template/internal/microservices/user/internal/repo"
	"github.com/gofrs/uuid"
)

func TestRepo_Smoke(t *testing.T) {
	t.Parallel()

	db, assert := start(t)

	m := metrics.DB("user", metrics.MethodsOf(&repo.Repo{})...)
	r := repo.New(db, &m)

	user := app.User{
		ID:        uuid.Nil,
		Email:     "email@gmail.com",
		Name:      "username",
		PassHash:  []byte("pass"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := user

	id, err := r.Save(ctx, user)
	assert.Nil(err)
	assert.NotNil(id)
	user.ID = id

	user.Name = "new_username"
	err = r.Update(ctx, user)
	assert.Nil(err)

	_, err = r.Save(ctx, user2)
	assert.ErrorIs(err, app.ErrEmailExist)

	user2.Email = "free@gmail.com"
	user2.Name = user.Name
	_, err = r.Save(ctx, user2)
	assert.ErrorIs(err, app.ErrUsernameExist)

	res, err := r.ByID(ctx, user.ID)
	assert.Nil(err)
	user.CreatedAt = res.CreatedAt
	user.UpdatedAt = res.UpdatedAt
	assert.Equal(user, *res)

	res, err = r.ByEmail(ctx, user.Email)
	assert.Nil(err)
	assert.Equal(user, *res)

	res, err = r.ByUsername(ctx, user.Name)
	assert.Nil(err)
	assert.Equal(user, *res)

	listRes, total, err := r.ListUserByUsername(ctx, user.Name, app.SearchParams{Limit: 5})
	assert.Nil(err)
	assert.Equal(1, total)
	assert.Equal([]app.User{user}, listRes)

	err = r.Delete(ctx, id)
	assert.Nil(err)

	res, err = r.ByID(ctx, user.ID)
	assert.Nil(res)
	assert.ErrorIs(err, app.ErrNotFound)
}
