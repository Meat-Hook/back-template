// Package web contains all methods for working web server.
package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"

	"github.com/Meat-Hook/back-template/internal/libs/log"
	"github.com/Meat-Hook/back-template/internal/libs/metrics"
	"github.com/Meat-Hook/back-template/internal/libs/middleware"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/api/web/generated/restapi/operations"
	"github.com/Meat-Hook/back-template/internal/modules/session/internal/app"
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
		Login(ctx context.Context, email, password string, origin app.Origin) (*app.User, *app.Token, error)
		Logout(ctx context.Context, session app.Session) error
		Session(ctx context.Context, accessToken string) (*app.Session, error)
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

	api := operations.NewSessionServiceAPI(swaggerSpec)
	swaggerLogger := logger.With().Str(log.Name, "swagger").Logger()
	api.Logger = swaggerLogger.Printf
	api.CookieKeyAuth = svc.cookieKeyAuth

	api.LoginHandler = operations.LoginHandlerFunc(svc.login)
	api.LogoutHandler = operations.LogoutHandlerFunc(svc.logout)

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
			Path:     "",
			SpecURL:  path.Join(swaggerSpec.BasePath(), "/swagger.json"),
			RedocURL: "",
			Title:    "",
		}

		return xffmw.Handler(createLog(middleware.Recovery(accesslog(middleware.Health(
			swag_middleware.Spec(swaggerSpec.BasePath(), restapi.FlatSwaggerJSON,
				swag_middleware.Redoc(redocOpts, handler)))))))
	}

	server.SetHandler(globalMiddlewares(api.Serve(nil)))

	return server, nil
}

func fromRequest(r *http.Request, session *app.Session) (context.Context, zerolog.Logger, net.IP) {
	ctx := r.Context()
	userID := 0
	if session != nil {
		userID = session.UserID
	}

	logger := zerolog.Ctx(r.Context()).With().Int(log.User, userID).Logger()
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	return ctx, logger, net.ParseIP(remoteIP)
}
