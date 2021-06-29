package web

import (
	models2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/models"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	"github.com/go-openapi/swag"
	"github.com/rs/zerolog"
)

func apiError(txt string) *models2.Error {
	return &models2.Error{
		Message: swag.String(txt),
	}
}

func logs(log zerolog.Logger, err error) {
	if err != nil {
		log.Error().Err(err).Send()
	}
}

// Users conversion []app.User => []*models.User.
func Users(u []app2.User) []*models2.User {
	users := make([]*models2.User, len(u))

	for i := range users {
		users[i] = User(&u[i])
	}

	return users
}

// User conversion app.User => models.User.
func User(u *app2.User) *models2.User {
	id := models2.UserID(u.ID.String())
	username := models2.Username(u.Name)
	email := models2.Email(u.Email)

	return &models2.User{
		ID:       &id,
		Username: &username,
		Email:    &email,
	}
}
