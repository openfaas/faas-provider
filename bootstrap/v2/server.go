// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

/*

OpenFaaS API server example:

import bootstrap "github.com/openfaas/faas-provider/bootstrap/v2"

config := bootstrap.ApiServerConfig{
	ReadTimeout:  cfg.ReadTimeout,
	WriteTimeout: cfg.WriteTimeout,
	TCPPort:      &port,
	EnableHealth: true,
}

handlers = bootstrap.ApiServerHandlers{
	FunctionProxy:  handlers.MakeProxy(functionNamespace, config.ReadTimeout),
	DeleteHandler:  handlers.MakeDeleteHandler(functionNamespace, clientset),
	DeployHandler:  handlers.MakeDeployHandler(functionNamespace, factory),
	FunctionReader: handlers.MakeFunctionReader(functionNamespace, clientset),
	ReplicaReader:  handlers.MakeReplicaReader(functionNamespace, clientset),
	ReplicaUpdater: handlers.MakeReplicaUpdater(functionNamespace, clientset),
	UpdateHandler:  handlers.MakeUpdateHandler(functionNamespace, factory),
	HealthHandler:  handlers.MakeHealthHandler(),
	InfoHandler:    handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommit),
	SecretHandler:  handlers.MakeSecretHandler(functionNamespace, clientset),
}

srv := bootstrap.NewApiServer(config, handlers, nil)

log.Fatal(srv.ListenAndServe())

*/

package v2

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/auth/v1"
)

// ApiServer represents the OpenFaaS API
type ApiServer struct {
	router   *mux.Router
	handlers ApiServerHandlers
	config   ApiServerConfig
}

// NewApiServer return an OpenFaaS API server
func NewApiServer(config ApiServerConfig, handlers ApiServerHandlers, router *mux.Router) *ApiServer {
	if router == nil {
		router = mux.NewRouter()
	}
	return &ApiServer{
		config:   config,
		handlers: handlers,
		router:   router,
	}
}

// ListenAndServe load your handlers into the correct OpenFaaS route spec and starts the API server.
// This function is blocking.
func (srv *ApiServer) ListenAndServe() error {
	if srv.config.EnableBasicAuth {
		reader := v1.ReadBasicAuthFromDisk{
			SecretMountPath: srv.config.SecretMountPath,
		}

		credentials, err := reader.Read()
		if err != nil {
			log.Fatal(err)
		}

		srv.handlers.FunctionReader = v1.DecorateWithBasicAuth(srv.handlers.FunctionReader, credentials)
		srv.handlers.DeployHandler = v1.DecorateWithBasicAuth(srv.handlers.DeployHandler, credentials)
		srv.handlers.DeleteHandler = v1.DecorateWithBasicAuth(srv.handlers.DeleteHandler, credentials)
		srv.handlers.UpdateHandler = v1.DecorateWithBasicAuth(srv.handlers.UpdateHandler, credentials)
		srv.handlers.ReplicaReader = v1.DecorateWithBasicAuth(srv.handlers.ReplicaReader, credentials)
		srv.handlers.ReplicaUpdater = v1.DecorateWithBasicAuth(srv.handlers.ReplicaUpdater, credentials)
		srv.handlers.InfoHandler = v1.DecorateWithBasicAuth(srv.handlers.InfoHandler, credentials)
		srv.handlers.SecretHandler = v1.DecorateWithBasicAuth(srv.handlers.SecretHandler, credentials)
	}

	// System (auth) endpoints
	srv.router.HandleFunc("/system/functions", srv.handlers.FunctionReader).Methods("GET")
	srv.router.HandleFunc("/system/functions", srv.handlers.DeployHandler).Methods("POST")
	srv.router.HandleFunc("/system/functions", srv.handlers.DeleteHandler).Methods("DELETE")
	srv.router.HandleFunc("/system/functions", srv.handlers.UpdateHandler).Methods("PUT")

	srv.router.HandleFunc("/system/function/{name:[-a-zA-Z_0-9]+}", srv.handlers.ReplicaReader).Methods("GET")
	srv.router.HandleFunc("/system/scale-function/{name:[-a-zA-Z_0-9]+}", srv.handlers.ReplicaUpdater).Methods("POST")
	srv.router.HandleFunc("/system/info", srv.handlers.InfoHandler).Methods("GET")

	srv.router.HandleFunc("/system/secrets", srv.handlers.SecretHandler).Methods(http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete)

	// Open endpoints
	srv.router.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}", srv.handlers.FunctionProxy)
	srv.router.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/", srv.handlers.FunctionProxy)
	srv.router.HandleFunc("/function/{name:[-a-zA-Z_0-9]+}/{params:.*}", srv.handlers.FunctionProxy)

	if srv.config.EnableHealth {
		srv.router.HandleFunc("/healthz", srv.handlers.Health).Methods("GET")
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", srv.config.TCPPort),
		ReadTimeout:    srv.config.ReadTimeout,
		WriteTimeout:   srv.config.WriteTimeout,
		MaxHeaderBytes: http.DefaultMaxHeaderBytes, // 1MB - can be overridden by setting Server.MaxHeaderBytes.
		Handler:        srv.router,
	}

	return s.ListenAndServe()
}
