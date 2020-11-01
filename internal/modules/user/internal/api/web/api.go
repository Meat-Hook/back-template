// Package web contains all methods for working web server.
package web

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/Meat-Hook/back-template/internal/libs/middleware"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/restapi/operations"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
	"github.com/go-openapi/loads"
	swag_middleware "github.com/go-openapi/runtime/middleware"
	"github.com/rs/zerolog"
	"github.com/sebest/xff"
)

//go:generate mockgen -source=api.go -destination mock.app.contracts_test.go -package web_test

type (
	// For easy testing.
	// Wrapper for app.Module.
	application interface {
		VerificationEmail(ctx context.Context, email string) error
		VerificationUsername(ctx context.Context, username string) error
		CreateUser(ctx context.Context, email string, username string, pass string) (int, error)
		UserByID(ctx context.Context, session app.Session, id int) (*app.User, error)
		DeleteUser(ctx context.Context, session app.Session) error
		ListUserByUsername(ctx context.Context, session app.Session, username string, page app.SearchParams) ([]app.User, int, error)
		UpdateUsername(ctx context.Context, session app.Session, username string) error
		UpdatePassword(ctx context.Context, session app.Session, oldPass string, newPass string) error
		Auth(ctx context.Context, raw string) (*app.Session, error)
	}

	service struct {
		app application
	}
	// Config for start server.
	Config struct {
		Host string
		Port int
	}
)

// New returns Swagger server configured to listen on the TCP network.
func New(module application, logger zerolog.Logger, m *metrics.API, cfg Config) (*restapi.Server, error) {
	svc := &service{
		app: module,
	}

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("load embedded swagger spec: %w", err)
	}

	api := operations.NewServiceUserAPI(swaggerSpec)
	swaggerLogger := logger.With().Str(log.Name, "swagger").Logger()
	api.Logger = swaggerLogger.Printf
	api.CookieKeyAuth = svc.cookieKeyAuth

	api.VerificationEmailHandler = operations.VerificationEmailHandlerFunc(svc.verificationEmail)
	api.VerificationUsernameHandler = operations.VerificationUsernameHandlerFunc(svc.verificationUsername)
	api.CreateUserHandler = operations.CreateUserHandlerFunc(svc.createUser)
	api.GetUserHandler = operations.GetUserHandlerFunc(svc.getUser)
	api.DeleteUserHandler = operations.DeleteUserHandlerFunc(svc.deleteUser)
	api.UpdatePasswordHandler = operations.UpdatePasswordHandlerFunc(svc.updatePassword)
	api.UpdateUsernameHandler = operations.UpdateUsernameHandlerFunc(svc.updateUsername)
	api.GetUsersHandler = operations.GetUsersHandlerFunc(svc.getUsers)

	server := restapi.NewServer(api)
	server.Host = cfg.Host
	server.Port = cfg.Port

	// The middlewareFunc executes before anything.
	globalMiddlewares := func(handler http.Handler) http.Handler {
		xffmw, _ := xff.Default()
		createLog := middleware.CreateLogger(logger.With())
		accesslog := middleware.AccessLog(m)
		redocOpts := swag_middleware.RedocOpts{
			BasePath: swaggerSpec.BasePath(),
			SpecURL:  path.Join(swaggerSpec.BasePath(), "/swagger.json"),
		}

		return xffmw.Handler(createLog(middleware.Recovery(accesslog(middleware.Health(
			swag_middleware.Spec(swaggerSpec.BasePath(), restapi.FlatSwaggerJSON,
				swag_middleware.Redoc(redocOpts, handler)))))))
	}

	server.SetHandler(globalMiddlewares(api.Serve(nil)))

	return server, nil
}

func fromRequest(r *http.Request, session *app.Session) (context.Context, zerolog.Logger) {
	ctx := r.Context()
	userID := 0
	if session != nil {
		userID = session.UserID
	}

	logger := zerolog.Ctx(r.Context()).With().Int(log.User, userID).Logger()

	return ctx, logger
}
