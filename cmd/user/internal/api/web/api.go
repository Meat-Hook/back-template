// Package web contains all methods for working web server.
package web

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/runtime"
	swag_middleware "github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/runtime/security"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"github.com/sebest/xff"

	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/restapi"
	"github.com/Meat-Hook/back-template/cmd/user/internal/api/web/generated/restapi/operations"
	"github.com/Meat-Hook/back-template/cmd/user/internal/app"
	"github.com/Meat-Hook/back-template/libs/log"
	"github.com/Meat-Hook/back-template/libs/web"
)

type (
	// For easy testing.
	// Wrapper for app.Module.
	application interface {
		VerificationEmail(ctx context.Context, email string) error
		VerificationUsername(ctx context.Context, username string) error
		CreateUser(ctx context.Context, email string, username string, pass string) (uuid.UUID, error)
		UserByID(ctx context.Context, session app.Session, id uuid.UUID) (*app.User, error)
		DeleteUser(ctx context.Context, session app.Session) error
		ListUserByUsername(ctx context.Context, session app.Session, username string, page app.SearchParams) ([]app.User, int, error)
		UpdateUsername(ctx context.Context, session app.Session, username string) error
		UpdatePassword(ctx context.Context, session app.Session, oldPass string, newPass string) error
		Login(ctx context.Context, email, password string, origin app.Origin) (*app.Token, error)
		Logout(ctx context.Context, session app.Session) error
		Auth(ctx context.Context, token string) (*app.Session, error)
		UploadAvatar(ctx context.Context, session app.Session, file io.Reader) error
		DeleteAvatar(ctx context.Context, session app.Session, fileID uuid.UUID) error
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
	svc := &service{
		app: module,
	}

	logger := zerolog.Ctx(ctx)

	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		return nil, fmt.Errorf("load embedded swagger spec: %w", err)
	}

	api := operations.NewUserServiceAPI(swaggerSpec)
	swaggerLogger := logger.With().Str(log.Subsystem, "swagger").Logger()
	api.Logger = swaggerLogger.Printf
	api.APIKeyAuthenticator = svc.authorizerFunc
	// Because it doesn't have context.
	// See previews code line.
	api.CookieKeyAuth = func(string) (*app.Session, error) {
		return nil, nil
	}

	api.VerificationEmailHandler = operations.VerificationEmailHandlerFunc(svc.verificationEmail)
	api.VerificationUsernameHandler = operations.VerificationUsernameHandlerFunc(svc.verificationUsername)
	api.CreateUserHandler = operations.CreateUserHandlerFunc(svc.createUser)
	api.GetUserHandler = operations.GetUserHandlerFunc(svc.getUser)
	api.DeleteUserHandler = operations.DeleteUserHandlerFunc(svc.deleteUser)
	api.UpdatePasswordHandler = operations.UpdatePasswordHandlerFunc(svc.updatePassword)
	api.UpdateUsernameHandler = operations.UpdateUsernameHandlerFunc(svc.updateUsername)
	api.GetUsersHandler = operations.GetUsersHandlerFunc(svc.getUsers)
	api.LoginHandler = operations.LoginHandlerFunc(svc.login)
	api.LogoutHandler = operations.LogoutHandlerFunc(svc.logout)
	api.NewAvatarHandler = operations.NewAvatarHandlerFunc(svc.uploadAvatar)
	api.DeleteAvatarHandler = operations.DeleteAvatarHandlerFunc(svc.deleteAvatar)

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
			SpecURL:  path.Join(swaggerSpec.BasePath(), "/swagger.json"),
		}

		return xffmw.Handler(createLog(web.Recovery(accesslog(web.Health(
			swag_middleware.Spec(swaggerSpec.BasePath(), restapi.FlatSwaggerJSON,
				swag_middleware.Redoc(redocOpts, handler)))))))
	}

	server.SetHandler(globalMiddlewares(api.Serve(nil)))
	return server, nil
}

func fromRequest(r *http.Request, session *app.Session) (context.Context, zerolog.Logger, net.IP) {
	ctx := r.Context()
	userID := uuid.Nil
	if session != nil {
		userID = session.UserID
	}

	logger := zerolog.Ctx(r.Context()).With().Stringer(log.User, userID).Logger()
	remoteIP, _, _ := net.SplitHostPort(r.RemoteAddr)

	return ctx, logger, net.ParseIP(remoteIP)
}

func generateCookie(token string) *http.Cookie {
	cookie := &http.Cookie{
		Name:       cookieTokenName,
		Value:      token,
		Path:       "/",
		Domain:     "",
		Expires:    time.Time{},
		RawExpires: "",
		MaxAge:     0,
		Secure:     true,
		HttpOnly:   true,
		SameSite:   http.SameSiteLaxMode,
		Raw:        "",
		Unparsed:   nil,
	}

	return cookie
}

func (s *service) authorizerFunc(name, in string, _ security.TokenAuthentication) runtime.Authenticator {
	const (
		query  = "query"
		header = "header"
	)

	inl := strings.ToLower(in)
	if inl != query && inl != header {
		// panic because this is most likely a typo
		panic(fmt.Sprintf(`api key auth: in value needs to be either "query" or "header"`))
	}

	var getToken func(*http.Request) string
	switch inl {
	case header:
		getToken = func(r *http.Request) string { return r.Header.Get(name) }
	case query:
		getToken = func(r *http.Request) string { return r.URL.Query().Get(name) }
	}

	return security.HttpAuthenticator(func(r *http.Request) (bool, interface{}, error) {
		token := getToken(r)
		if token == "" {
			return false, nil, nil
		}

		p, err := s.cookieKeyAuth(r.Context(), token)
		return true, p, err
	})
}
