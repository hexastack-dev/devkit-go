// Copyright 2018 The Go Cloud Development Kit Authors
// Copyright 2023 hexastack.dev
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// This file has been modified by hexastack.dev to meet our development
// requirements, including but not limited to:
// - remove wire dependency
// - use our server package instead of gocloud.dev/server
// - use personalized errors
// - use OpenTelemetry instead of OpenCensus
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package server provides a preconfigured HTTP server with diagnostic hooks.
package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/hexastack-dev/devkit-go/log"
	"github.com/hexastack-dev/devkit-go/server/driver"
	"github.com/hexastack-dev/devkit-go/server/health"
	"github.com/hexastack-dev/devkit-go/server/recoverer"
	"github.com/hexastack-dev/devkit-go/server/requestlog"
	"github.com/hexastack-dev/devkit-go/server/rwlog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// Server is a preconfigured HTTP server with diagnostic hooks.
// The zero value is a server with the default options.
type Server struct {
	reqlog         requestlog.Logger
	handler        http.Handler
	wrappedHandler http.Handler
	healthHandler  health.Handler
	once           sync.Once
	driver         driver.Server

	logger log.Logger
}

// Options is the set of optional parameters.
type Options struct {
	// RequestLogger specifies the logger that will be used to log requests.
	RequestLogger requestlog.Logger

	// HealthChecks specifies the health checks to be run when the
	// /healthz/readiness endpoint is requested.
	HealthChecks []health.Checker

	// Driver serves HTTP requests.
	Driver driver.Server

	// PanicHandler specifies http.Handler that will be called when panic occured.
	// If nil, then default PanicHandler will be used.
	PanicHandler http.Handler

	// Logger specifies logger to use by Server when specific events occurs.
	Logger log.Logger
}

// New creates a new server. New(nil, nil) is the same as new(Server).
func New(h http.Handler, opts *Options) *Server {
	srv := &Server{handler: h}
	var panicHandler http.Handler
	if opts != nil {
		srv.reqlog = opts.RequestLogger
		for _, c := range opts.HealthChecks {
			srv.healthHandler.Add(c)
		}
		srv.driver = opts.Driver

		srv.logger = opts.Logger
		if opts.PanicHandler != nil {
			panicHandler = opts.PanicHandler
		}
	}

	srv.handler = recoverer.New(panicHandler)(srv.handler)

	return srv
}

func (srv *Server) init() {
	srv.once.Do(func() {
		if srv.driver == nil {
			srv.driver = NewDefaultDriver()
		}
		if srv.handler == nil {
			srv.handler = http.DefaultServeMux
		}
		// Setup health checks, /healthz route is taken by health checks by default.
		// Note: App Engine Flex uses /_ah/health by default, which can be changed
		// in app.yaml. We may want to do an auto-detection for flex in future.
		const healthPrefix = "/healthz/"

		mux := http.NewServeMux()
		mux.HandleFunc(healthPrefix+"liveness", health.HandleLive)
		mux.Handle(healthPrefix+"readiness", &srv.healthHandler)
		h := srv.handler
		if srv.reqlog != nil {
			h = requestlog.New(srv.reqlog)(h)
		}
		h = rwlog.New(func(err error) {
			getLogger(srv.logger).Error("Error when writing response", err)
		})(h)

		// h = otelhttp.NewHandler(h, os.Args[0])
		h = otelhttp.NewHandler(h, "")
		mux.Handle("/", h)
		srv.wrappedHandler = mux
	})
}

func getLogger(logger log.Logger) log.Logger {
	if logger == nil {
		return log.GetLogger()
	}
	return logger
}

// ListenAndServe is a wrapper to use wherever http.ListenAndServe is used.
// It wraps the http.Handler provided to New with a handler that handles tracing and
// request logging. If the handler is nil, then http.DefaultServeMux will be used.
// A configured Requestlogger will log all requests except HealthChecks.
func (srv *Server) ListenAndServe(addr string) error {
	srv.init()
	getLogger(srv.logger).Debug("Listen and serve at: " + addr)
	return srv.driver.ListenAndServe(addr, srv.wrappedHandler)
}

// ListenAndServeTLS is a wrapper to use wherever http.ListenAndServeTLS is used.
// It wraps the http.Handler provided to New with a handler that handles tracing and
// request logging. If the handler is nil, then http.DefaultServeMux will be used.
// A configured Requestlogger will log all requests except HealthChecks.
func (srv *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	// Check if the driver implements the optional interface.
	tlsDriver, ok := srv.driver.(driver.TLSServer)
	if !ok {
		return fmt.Errorf("driver %T does not support ListenAndServeTLS", srv.driver)
	}
	srv.init()
	getLogger(srv.logger).Debug("Listen and serve TLS at: " + addr)
	return tlsDriver.ListenAndServeTLS(addr, certFile, keyFile, srv.wrappedHandler)
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (srv *Server) Shutdown(ctx context.Context) error {
	if srv.driver == nil {
		return nil
	}
	return srv.driver.Shutdown(ctx)
}

// DefaultDriver implements the driver.Server interface. The zero value is a valid http.Server.
type DefaultDriver struct {
	Server http.Server
}

// NewDefaultDriver creates a driver with an http.Server with default timeouts.
func NewDefaultDriver() *DefaultDriver {
	return &DefaultDriver{
		Server: http.Server{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

// ListenAndServe sets the address and handler on DefaultDriver's http.Server,
// then calls ListenAndServe on it.
func (dd *DefaultDriver) ListenAndServe(addr string, h http.Handler) error {
	dd.Server.Addr = addr
	dd.Server.Handler = h
	return dd.Server.ListenAndServe()
}

// ListenAndServeTLS sets the address and handler on DefaultDriver's http.Server,
// then calls ListenAndServeTLS on it.
//
// DefaultDriver.Server.TLSConfig may be set to configure additional TLS settings.
func (dd *DefaultDriver) ListenAndServeTLS(addr, certFile, keyFile string, h http.Handler) error {
	dd.Server.Addr = addr
	dd.Server.Handler = h
	return dd.Server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown gracefully shuts down the server without interrupting any active connections,
// by calling Shutdown on DefaultDriver's http.Server
func (dd *DefaultDriver) Shutdown(ctx context.Context) error {
	return dd.Server.Shutdown(ctx)
}
