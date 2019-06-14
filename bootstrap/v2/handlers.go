// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v2

import "net/http"

// ApiServerHandlers provide handlers for OpenFaaS API
type ApiServerHandlers struct {
	FunctionReader http.HandlerFunc
	DeployHandler  http.HandlerFunc
	DeleteHandler  http.HandlerFunc
	ReplicaReader  http.HandlerFunc
	ReplicaUpdater http.HandlerFunc
	SecretHandler  http.HandlerFunc

	// FunctionProxy provides the function invocation proxy logic. Use proxy.NewHandlerFunc to
	// use the standard OpenFaaS proxy implementation or provide completely custom proxy logic.
	FunctionProxy http.HandlerFunc

	// Optional: Update an existing function
	UpdateHandler http.HandlerFunc
	Health        http.HandlerFunc
	InfoHandler   http.HandlerFunc
}
