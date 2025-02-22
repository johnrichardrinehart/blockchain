// Package handlers manages the different versions of the API.
package handlers

import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/ardanlabs/blockchain/app/services/node/handlers/debug/checkgrp"
	v1 "github.com/ardanlabs/blockchain/app/services/node/handlers/v1"
	"github.com/ardanlabs/blockchain/business/web/v1/mid"
	"github.com/ardanlabs/blockchain/foundation/blockchain"
	"github.com/ardanlabs/blockchain/foundation/web"
	"go.uber.org/zap"
)

// MuxConfig contains all the mandatory systems required by handlers.
type MuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	BC       *blockchain.State
}

// PublicMux constructs a http.Handler with all application routes defined.
func PublicMux(cfg MuxConfig) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	// Load the v1 routes.
	v1.PublicRoutes(app, v1.Config{
		Log: cfg.Log,
		BC:  cfg.BC,
	})

	return app
}

// PrivateMux constructs a http.Handler with all application routes defined.
func PrivateMux(cfg MuxConfig) http.Handler {

	// Construct the web.App which holds all routes as well as common Middleware.
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	// Load the v1 routes.
	v1.PrivateRoutes(app, v1.Config{
		Log: cfg.Log,
		BC:  cfg.BC,
	})

	return app
}

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux. Using the
// DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

// DebugMux registers all the debug standard library routes and then custom
// debug application routes for the service. This bypassing the use of the
// DefaultServerMux. Using the DefaultServerMux would be a security risk since
// a dependency could inject a handler into our service without us knowing it.
func DebugMux(build string, log *zap.SugaredLogger) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug check endpoints.
	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
	}
	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}
