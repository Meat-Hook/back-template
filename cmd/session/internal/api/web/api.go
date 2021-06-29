// Package web contains all methods for working web server.
package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"

	restapi2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web/generated/restapi"
	operations2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/api/web/generated/restapi/operations"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/session/internal/app"
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
		Login(ctx context.Context, email, password string, origin app2.Origin) (*app2.User, *app2.Token, error)
		Logout(ctx context.Context, session app2.Session) error
		Session(ctx context.Context, accessToken string) (*app2.Session, error)
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

	api := operations2.NewSessionServiceAPI(swaggerSpec)
	swaggerLogger := logger.With().Str(log2.Name, "swagger").Logger()
	api.Logger = swaggerLogger.Printf
	api.CookieKeyAuth = svc.cookieKeyAuth

	api.LoginHandler = operations2.LoginHandlerFunc(svc.login)
	api.LogoutHandler = operations2.LogoutHandlerFunc(svc.logout)

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
			Path:     "",
			SpecURL:  path.Join(swaggerSpec.BasePath(), "/swagger.json"),
			RedocURL: "",
			Title:    "",
		}

		return xffmw.Handler(createLog(middleware2.Recovery(accesslog(middleware2.Health(
			swag_middleware.Spec(swaggerSpec.BasePath(), restapi2.FlatSwaggerJSON,
				swag_middleware.Redoc(redocOpts, handler)))))))
	}

	server.SetHandler(globalMiddlewares(api.Serve(nil)))

	return server, nil
}

func fromRequest(r *http.Request, session *app2.Session) (context.Context, zerolog.Logger, net.IP) {
	ctx := r.Context()
	userID := uuid.Nil
	if session != nil {
		userID = session.UserID
	}

	logger := zerolog.Ctx(r.Context()).With().Stringer(log2.User, userID).Logger()
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	return ctx, logger, net.ParseIP(remoteIP)
}
