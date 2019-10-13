package types

import (
	"net/http"
	"time"
)

// FaaSHandlers provide handlers for OpenFaaS
type FaaSHandlers struct {
	// FunctionProxy provides the function invocation proxy logic.  Use proxy.NewHandlerFunc to
	// use the standard OpenFaaS proxy implementation or provide completely custom proxy logic.
	FunctionProxy http.HandlerFunc

	FunctionReader http.HandlerFunc
	DeployHandler  http.HandlerFunc

	DeleteHandler  http.HandlerFunc
	ReplicaReader  http.HandlerFunc
	ReplicaUpdater http.HandlerFunc
	SecretHandler  http.HandlerFunc
	// LogHandler provides streaming json logs of functions
	LogHandler http.HandlerFunc

	// UpdateHandler an existing function/service
	UpdateHandler        http.HandlerFunc
	HealthHandler        http.HandlerFunc
	InfoHandler          http.HandlerFunc
	ListNamespaceHandler http.HandlerFunc
}

// FaaSConfig set config for HTTP handlers
type FaaSConfig struct {
	// TCPPort is the public port for the API.
	TCPPort *int
	// HTTP timeout for reading a request from clients.
	ReadTimeout time.Duration
	// HTTP timeout for writing a response from functions.
	WriteTimeout time.Duration
	// EnableHealth enables/disables the default health endpoint bound to "/healthz".
	EnableHealth bool
	// EnableBasicAuth enforces basic auth on the API. If set, reads secrets from file-system
	// location specificed in `SecretMountPath`.
	EnableBasicAuth bool
	// SecretMountPath specifies where to read secrets from for embedded basic auth.
	SecretMountPath string
	// MaxIdleConns with a default value of 1024, can be used for tuning HTTP proxy performance.
	MaxIdleConns int
	// MaxIdleConnsPerHost with a default value of 1024, can be used for tuning HTTP proxy performance.
	MaxIdleConnsPerHost int
}
