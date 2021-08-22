// Package web contains all methods for working web server.
package web

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"

	"github.com/go-openapi/loads"
	swag_middleware "github.com/go-openapi/runtime/middleware"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"github.com/sebest/xff"

	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/restapi/operations"
	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/web"
)

//go:generate mockgen -source=api.go -destination mock.app.contracts_test.go -package web_test

type (
	// For easy testing.
	// Wrapper for app.Module.
	application interface {
		GetFile(ctx context.Context, fileID uuid.UUID) (*app.File, error)
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
func New(ctx context.Context, module application, m *web.Metric, cfg Config) (*restapi.Server, error) {
	logger := zerolog.Ctx(ctx)
	svc := &service{
		app: module,
	}

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("load embedded swagger spec: %w", err)
	}

	api := operations.NewFileServiceAPI(swaggerSpec)
	swaggerLogger := logger.With().Str(log.Subsystem, "swagger").Logger()
	api.Logger = swaggerLogger.Printf

	api.GetFileHandler = operations.GetFileHandlerFunc(svc.getFile)

	server := restapi.NewServer(api)
	server.Host = cfg.Host
	server.Port = cfg.Port

	// The middlewareFunc executes before anything.
	globalMiddlewares := func(handler http.Handler) http.Handler {
		xffmw, _ := xff.Default()
		createLog := web.CreateLogger(logger.With())
		accesslog := web.AccessLog(m)
		redocOpts := swag_middleware.RedocOpts{
			BasePath: swaggerSpec.BasePath(),
			Path:     "",
			SpecURL:  path.Join(swaggerSpec.BasePath(), "/swagger.json"),
			RedocURL: "",
			Title:    "",
		}

		return xffmw.Handler(createLog(web.Recovery(accesslog(web.Health(
			swag_middleware.Spec(swaggerSpec.BasePath(), restapi.FlatSwaggerJSON,
				swag_middleware.Redoc(redocOpts, handler)))))))
	}

	server.SetHandler(globalMiddlewares(api.Serve(nil)))

	return server, nil
}

func fromRequest(r *http.Request) (context.Context, zerolog.Logger, net.IP) {
	ctx := r.Context()

	logger := zerolog.Ctx(r.Context())
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	return ctx, *logger, net.ParseIP(remoteIP)
}
