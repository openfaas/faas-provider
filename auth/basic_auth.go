// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package auth

import (
	"net/http"
)

// DecorateWithBasicAuth enforces basic auth as a middleware with given credentials
func DecorateWithBasicAuth(next http.HandlerFunc, credentials *BasicAuthCredentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		user, password, ok := r.BasicAuth()
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		if !ok || !(credentials.Password == password && user == credentials.User) {
			unauthorized(w, r)
		}

		next.ServeHTTP(w, r)
	}
}

// Support for Multiple User Basic Auth Providers
func DecorateWithMultiUserBasicAuth(next http.HandlerFunc, creds []BasicAuthCredentials) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ok := false
		user, password, _ := r.BasicAuth()
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		for _, v := range creds {
			if v.User == user && v.Password == password {
				ok = true
				break
			}
		}

		if !ok {
			unauthorized(w, r)
		}

		next.ServeHTTP(w, r)
	}
}

// Helper function to return a 401
func unauthorized(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("invalid credentials"))
	return
}
