package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/openshift-kni/oran-hwmgr-plugin/internal/server/api/generated"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	oapimiddleware "github.com/oapi-codegen/nethttp-middleware"
)

type Middleware = func(http.Handler) http.Handler

// GetLogDurationFunc log time taken to complete a request.
func GetLogDurationFunc() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			next.ServeHTTP(w, r)
			slog.Debug(fmt.Sprintf("%s took %s", r.RequestURI, time.Since(startTime)))
		})
	}
}

// GetOpenAPIValidationFunc to validate all incoming requests as specified in the spec
func GetOpenAPIValidationFunc(swagger *openapi3.T) Middleware {
	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	return oapimiddleware.OapiRequestValidatorWithOptions(swagger, &oapimiddleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc, // No auth needed even when we have something in spec
		},
		ErrorHandler: getErrorHandlerFunc(),
	})
}

// problemDetails writes an error message using the appropriate header for an ORAN error response
func problemDetails(w http.ResponseWriter, body string, code int) {
	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(code)
	_, err := fmt.Fprintln(w, body)
	if err != nil {
		panic(err)
	}
}

// getErrorHandlerFunc override default validation error to allow for O-RAN specific error
func getErrorHandlerFunc() func(w http.ResponseWriter, message string, statusCode int) {
	return func(w http.ResponseWriter, message string, statusCode int) {
		out, _ := json.Marshal(generated.ProblemDetails{
			Detail: message,
			Status: statusCode,
		})
		problemDetails(w, string(out), statusCode)
	}
}

// GracefulShutdown allow graceful shutdown with timeout
func GracefulShutdown(srv *http.Server) error {
	// Create shutdown context with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed graceful shutdown: %w", err)
	}

	slog.Info("Server gracefully stopped")
	return nil
}

// GetRequestErrorFunc override default validation errors to allow for O-RAN specific struct
func GetRequestErrorFunc() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		out, _ := json.Marshal(generated.ProblemDetails{
			Detail: err.Error(),
			Status: http.StatusBadRequest,
		})
		problemDetails(w, string(out), http.StatusBadRequest)
	}
}

// GetResponseErrorFunc override default internal server error to allow for O-RAN specific struct
func GetResponseErrorFunc() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		out, _ := json.Marshal(generated.ProblemDetails{
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		})
		problemDetails(w, string(out), http.StatusInternalServerError)
	}
}

// GetNotFoundFunc is used to override the default 404 response which is a text only reply so that we can respond with
// the required JSON body.
func GetNotFoundFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		out, _ := json.Marshal(generated.ProblemDetails{
			Detail: fmt.Sprintf("Path '%s' not found", r.RequestURI),
			Status: http.StatusNotFound,
		})
		problemDetails(w, string(out), http.StatusNotFound)
	}
}
