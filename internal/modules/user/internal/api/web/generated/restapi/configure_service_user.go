// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"

	"github.com/Meat-Hook/back-template/internal/modules/user/internal/api/web/generated/restapi/operations"
	"github.com/Meat-Hook/back-template/internal/modules/user/internal/app"
)

//go:generate swagger generate server --target ../../generated --name ServiceUser --spec ../../../../../../../../../../../../../../var/folders/8m/ql24vlxd54x78711p7fm6zvr0000gn/T/swagger.yml806659975 --principal github.com/Meat-Hook/back-template/internal/modules/user/internal/app.Session --exclude-main --strict-responders

func configureFlags(api *operations.ServiceUserAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.ServiceUserAPI) http.Handler {
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
		api.CookieKeyAuth = func(token string) (*app.Session, error) {
			return nil, errors.NotImplemented("api key auth (cookieKey) Cookie from header param [Cookie] has not yet been implemented")
		}
	}

	// Set your custom authorizer if needed. Default one is security.Authorized()
	// Expected interface runtime.Authorizer
	//
	// Example:
	// api.APIAuthorizer = security.Authorized()
	if api.CreateUserHandler == nil {
		api.CreateUserHandler = operations.CreateUserHandlerFunc(func(params operations.CreateUserParams) operations.CreateUserResponder {
			return operations.CreateUserNotImplemented()
		})
	}
	if api.DeleteUserHandler == nil {
		api.DeleteUserHandler = operations.DeleteUserHandlerFunc(func(params operations.DeleteUserParams, principal *app.Session) operations.DeleteUserResponder {
			return operations.DeleteUserNotImplemented()
		})
	}
	if api.GetUserHandler == nil {
		api.GetUserHandler = operations.GetUserHandlerFunc(func(params operations.GetUserParams, principal *app.Session) operations.GetUserResponder {
			return operations.GetUserNotImplemented()
		})
	}
	if api.GetUsersHandler == nil {
		api.GetUsersHandler = operations.GetUsersHandlerFunc(func(params operations.GetUsersParams, principal *app.Session) operations.GetUsersResponder {
			return operations.GetUsersNotImplemented()
		})
	}
	if api.UpdatePasswordHandler == nil {
		api.UpdatePasswordHandler = operations.UpdatePasswordHandlerFunc(func(params operations.UpdatePasswordParams, principal *app.Session) operations.UpdatePasswordResponder {
			return operations.UpdatePasswordNotImplemented()
		})
	}
	if api.UpdateUsernameHandler == nil {
		api.UpdateUsernameHandler = operations.UpdateUsernameHandlerFunc(func(params operations.UpdateUsernameParams, principal *app.Session) operations.UpdateUsernameResponder {
			return operations.UpdateUsernameNotImplemented()
		})
	}
	if api.VerificationEmailHandler == nil {
		api.VerificationEmailHandler = operations.VerificationEmailHandlerFunc(func(params operations.VerificationEmailParams) operations.VerificationEmailResponder {
			return operations.VerificationEmailNotImplemented()
		})
	}
	if api.VerificationUsernameHandler == nil {
		api.VerificationUsernameHandler = operations.VerificationUsernameHandlerFunc(func(params operations.VerificationUsernameParams) operations.VerificationUsernameResponder {
			return operations.VerificationUsernameNotImplemented()
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
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
