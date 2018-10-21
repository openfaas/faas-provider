package types

import (
	"net/http"
	"time"

	"github.com/openfaas/faas-provider/proxy"
)

// FaaSHandlers provide handlers for OpenFaaS
type FaaSHandlers struct {
	FunctionReader http.HandlerFunc
	DeployHandler  http.HandlerFunc
	DeleteHandler  http.HandlerFunc
	ReplicaReader  http.HandlerFunc
	ReplicaUpdater http.HandlerFunc

	// Optional: Update an existing function
	UpdateHandler http.HandlerFunc
	Health        http.HandlerFunc
	InfoHandler   http.HandlerFunc
}

// FaaSConfig set config for HTTP handlers
type FaaSConfig struct {
	TCPPort         *int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	EnableHealth    bool
	EnableBasicAuth bool
	SecretMountPath string

	// The FaaS provider implementation is responsible for providing the resolver function implementation.
	// BaseURLResolver.Resolve will receive the function name and should return the base Address of the
	// function service.
	FunctionProxyResolver proxy.BaseURLResolver
}
