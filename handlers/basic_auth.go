// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package handlers

import (
	"net/http"

	"github.com/openfaas/faas-provider/types"
)

// DecorateWithBasicAuth enforces basic auth as a middleware with given credentials
func DecorateWithBasicAuth(next http.HandlerFunc, credentials *types.BasicAuthCredentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, password, ok := r.BasicAuth()
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		if !ok || credentials.Password != password || credentials.User != user {

			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("invalid credentials"))
			return
		}

		next.ServeHTTP(w, r)
	}
}
