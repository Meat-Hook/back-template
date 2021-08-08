package web

import (
	"github.com/go-openapi/strfmt"

	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/cmd/user/internal/app"

	"github.com/go-openapi/swag"
	"github.com/rs/zerolog"
)

func apiError(txt string) *models.Error {
	return &models.Error{
		Message: swag.String(txt),
	}
}

func logs(log zerolog.Logger, err error) {
	if err != nil {
		log.Error().Err(err).Send()
	}
}

// Users conversion []app.User => []*models.User.
func Users(u []app.User) []*models.User {
	users := make([]*models.User, len(u))

	for i := range users {
		users[i] = User(&u[i])
	}

	return users
}

// User conversion app.User => models.User.
func User(u *app.User) *models.User {
	id := models.UserID(u.ID.String())
	username := models.Username(u.Name)
	email := models.Email(u.Email)

	avatars := make([]strfmt.UUID, len(u.Avatars))
	for i := range u.Avatars {
		avatars[i] = strfmt.UUID(u.Avatars[i].String())
	}

	return &models.User{
		ID:       &id,
		Username: &username,
		Email:    &email,
		Avatars:  avatars,
	}
}
