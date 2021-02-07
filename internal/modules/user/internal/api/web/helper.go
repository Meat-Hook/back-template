package web

import (
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
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
	return &models.User{
		ID:       models.UserID(u.ID),
		Username: models.Username(u.Name),
		Email:    models.Email(u.Email),
	}
}
