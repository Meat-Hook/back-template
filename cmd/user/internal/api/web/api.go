// Package web contains all methods for working web server.
package web

import (
	"context"
	"fmt"
	"net/http"
	"path"

	restapi2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/restapi"
	operations2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/restapi/operations"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	log2 "github.com/Meat-Hook/back-template/internal/libs/log"
	metrics2 "github.com/Meat-Hook/back-template/internal/libs/metrics"
	middleware2 "github.com/Meat-Hook/back-template/internal/libs/middleware"
	"github.com/go-openapi/loads"
	swag_middleware "github.com/go-openapi/runtime/middleware"
	"github.com/gofrs/uuid"
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
		CreateUser(ctx context.Context, email string, username string, pass string) (uuid.UUID, error)
		UserByID(ctx context.Context, session app2.Session, id uuid.UUID) (*app2.User, error)
		DeleteUser(ctx context.Context, session app2.Session) error
		ListUserByUsername(ctx context.Context, session app2.Session, username string, page app2.SearchParams) ([]app2.User, int, error)
		UpdateUsername(ctx context.Context, session app2.Session, username string) error
		UpdatePassword(ctx context.Context, session app2.Session, oldPass string, newPass string) error
		Auth(ctx context.Context, raw string) (*app2.Session, error)
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
func New(module application, logger zerolog.Logger, m *metrics2.API, cfg Config) (*restapi2.Server, error) {
	svc := &service{
		app: module,
	}

	swaggerSpec, err := loads.Embedded(restapi2.SwaggerJSON, restapi2.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("load embedded swagger spec: %w", err)
	}

	api := operations2.NewUserServiceAPI(swaggerSpec)
	swaggerLogger := logger.With().Str(log2.Name, "swagger").Logger()
	api.Logger = swaggerLogger.Printf
	api.CookieKeyAuth = svc.cookieKeyAuth

	api.VerificationEmailHandler = operations2.VerificationEmailHandlerFunc(svc.verificationEmail)
	api.VerificationUsernameHandler = operations2.VerificationUsernameHandlerFunc(svc.verificationUsername)
	api.CreateUserHandler = operations2.CreateUserHandlerFunc(svc.createUser)
	api.GetUserHandler = operations2.GetUserHandlerFunc(svc.getUser)
	api.DeleteUserHandler = operations2.DeleteUserHandlerFunc(svc.deleteUser)
	api.UpdatePasswordHandler = operations2.UpdatePasswordHandlerFunc(svc.updatePassword)
	api.UpdateUsernameHandler = operations2.UpdateUsernameHandlerFunc(svc.updateUsername)
	api.GetUsersHandler = operations2.GetUsersHandlerFunc(svc.getUsers)

	server := restapi2.NewServer(api)
	server.Host = cfg.Host
	server.Port = cfg.Port

	// The middlewareFunc executes before anything.
	globalMiddlewares := func(handler http.Handler) http.Handler {
		xffmw, _ := xff.Default()
		createLog := middleware2.CreateLogger(logger.With())
		accesslog := middleware2.AccessLog(m)
		redocOpts := swag_middleware.RedocOpts{
			BasePath: swaggerSpec.BasePath(),
			SpecURL:  path.Join(swaggerSpec.BasePath(), "/swagger.json"),
		}

		return xffmw.Handler(createLog(middleware2.Recovery(accesslog(middleware2.Health(
			swag_middleware.Spec(swaggerSpec.BasePath(), restapi2.FlatSwaggerJSON,
				swag_middleware.Redoc(redocOpts, handler)))))))
	}

	server.SetHandler(globalMiddlewares(api.Serve(nil)))

	return server, nil
}

func fromRequest(r *http.Request, session *app2.Session) (context.Context, zerolog.Logger) {
	ctx := r.Context()
	userID := uuid.Nil
	if session != nil {
		userID = session.UserID
	}

	logger := zerolog.Ctx(r.Context()).With().Stringer(log2.User, userID).Logger()

	return ctx, logger
}
