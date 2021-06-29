// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	operations2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/api/web/generated/restapi/operations"
	app2 "github.com/Meat-Hook/back-template/internal/cmd/user/internal/app"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
)

//go:generate swagger generate server --target ../../generated --name UserService --spec ../../../../../swagger.yml --principal github.com/Meat-Hook/back-template/cmd/user/internal/app.Session --exclude-main --strict-responders

func configureFlags(api *operations2.UserServiceAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations2.UserServiceAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	// Applies when the "Cookie" header is set
	if api.CookieKeyAuth == nil {
		api.CookieKeyAuth = func(token string) (*app2.Session, error) {
			return nil, errors.NotImplemented("api key auth (cookieKey) Cookie from header param [Cookie] has not yet been implemented")
		}
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()

	if api.CreateUserHandler == nil {
		api.CreateUserHandler = operations2.CreateUserHandlerFunc(func(params operations2.CreateUserParams) operations2.CreateUserResponder {
			return operations2.CreateUserNotImplemented()
		})
	}
	if api.DeleteUserHandler == nil {
		api.DeleteUserHandler = operations2.DeleteUserHandlerFunc(func(params operations2.DeleteUserParams, principal *app2.Session) operations2.DeleteUserResponder {
			return operations2.DeleteUserNotImplemented()
		})
	}
	if api.GetUserHandler == nil {
		api.GetUserHandler = operations2.GetUserHandlerFunc(func(params operations2.GetUserParams, principal *app2.Session) operations2.GetUserResponder {
			return operations2.GetUserNotImplemented()
		})
	}
	if api.GetUsersHandler == nil {
		api.GetUsersHandler = operations2.GetUsersHandlerFunc(func(params operations2.GetUsersParams, principal *app2.Session) operations2.GetUsersResponder {
			return operations2.GetUsersNotImplemented()
		})
	}
	if api.UpdatePasswordHandler == nil {
		api.UpdatePasswordHandler = operations2.UpdatePasswordHandlerFunc(func(params operations2.UpdatePasswordParams, principal *app2.Session) operations2.UpdatePasswordResponder {
			return operations2.UpdatePasswordNotImplemented()
		})
	}
	if api.UpdateUsernameHandler == nil {
		api.UpdateUsernameHandler = operations2.UpdateUsernameHandlerFunc(func(params operations2.UpdateUsernameParams, principal *app2.Session) operations2.UpdateUsernameResponder {
			return operations2.UpdateUsernameNotImplemented()
		})
	}
	if api.VerificationEmailHandler == nil {
		api.VerificationEmailHandler = operations2.VerificationEmailHandlerFunc(func(params operations2.VerificationEmailParams) operations2.VerificationEmailResponder {
			return operations2.VerificationEmailNotImplemented()
		})
	}
	if api.VerificationUsernameHandler == nil {
		api.VerificationUsernameHandler = operations2.VerificationUsernameHandlerFunc(func(params operations2.VerificationUsernameParams) operations2.VerificationUsernameResponder {
			return operations2.VerificationUsernameNotImplemented()
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
