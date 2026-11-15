package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
	"github.com/prometheus/client_golang/prometheus"
)

func TestDecorateNamespaceHandlersWithBasicAuth(t *testing.T) {
	credentials := &auth.BasicAuthCredentials{
		User:     "admin",
		Password: "correct-horse-battery-staple",
	}

	tests := []struct {
		name    string
		method  string
		path    string
		handler func(*types.FaaSHandlers) http.HandlerFunc
	}{
		{
			name:   "list namespaces",
			method: http.MethodGet,
			path:   "/system/namespaces",
			handler: func(handlers *types.FaaSHandlers) http.HandlerFunc {
				return handlers.ListNamespaces
			},
		},
		{
			name:   "mutate namespace",
			method: http.MethodPost,
			path:   "/system/namespace/",
			handler: func(handlers *types.FaaSHandlers) http.HandlerFunc {
				return handlers.MutateNamespace
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := 0
			stub := func(w http.ResponseWriter, _ *http.Request) {
				called++
				w.WriteHeader(http.StatusAccepted)
			}

			handlers := types.FaaSHandlers{
				ListNamespaces:  stub,
				MutateNamespace: stub,
			}
			decorateNamespaceHandlersWithBasicAuth(&handlers, credentials)

			unauthenticated := httptest.NewRequest(tc.method, tc.path, nil)
			unauthenticatedResponse := httptest.NewRecorder()
			tc.handler(&handlers)(unauthenticatedResponse, unauthenticated)

			if unauthenticatedResponse.Code != http.StatusUnauthorized {
				t.Fatalf("want unauthenticated status %d, got %d", http.StatusUnauthorized, unauthenticatedResponse.Code)
			}
			if called != 0 {
				t.Fatalf("want protected handler not to be called, got %d calls", called)
			}

			authenticated := httptest.NewRequest(tc.method, tc.path, nil)
			authenticated.SetBasicAuth(credentials.User, credentials.Password)
			authenticatedResponse := httptest.NewRecorder()
			tc.handler(&handlers)(authenticatedResponse, authenticated)

			if authenticatedResponse.Code != http.StatusAccepted {
				t.Fatalf("want authenticated status %d, got %d", http.StatusAccepted, authenticatedResponse.Code)
			}
			if called != 1 {
				t.Fatalf("want protected handler to be called once, got %d calls", called)
			}
		})
	}
}

func TestDecorateNamespaceHandlersWithBasicAuthAllowsNilHandlers(t *testing.T) {
	handlers := types.FaaSHandlers{}
	decorateNamespaceHandlersWithBasicAuth(&handlers, &auth.BasicAuthCredentials{})
}

func TestMakeHandlerIsIsolated(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{name: "first handler", status: http.StatusAccepted},
		{name: "second handler", status: http.StatusCreated},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handlers := types.FaaSHandlers{
				ListNamespaces: func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(tc.status)
				},
			}
			registry := prometheus.NewRegistry()
			server := httptest.NewServer(MakeHandler(&handlers, &types.FaaSConfig{}, registry, registry))
			t.Cleanup(server.Close)

			response, err := server.Client().Get(server.URL + "/system/namespaces")
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != tc.status {
				t.Fatalf("want status %d, got %d", tc.status, response.StatusCode)
			}
		})
	}
}

func TestMakeHandlerProtectsNilNamespaceLister(t *testing.T) {
	const (
		user     = "admin"
		password = "correct-horse-battery-staple"
	)

	secretMountPath := t.TempDir()
	for name, value := range map[string]string{
		"basic-auth-user":     user,
		"basic-auth-password": password,
	} {
		if err := os.WriteFile(filepath.Join(secretMountPath, name), []byte(value), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}

	handlers := types.FaaSHandlers{}
	config := types.FaaSConfig{
		EnableBasicAuth: true,
		SecretMountPath: secretMountPath,
	}
	registry := prometheus.NewRegistry()
	server := httptest.NewServer(MakeHandler(&handlers, &config, registry, registry))
	t.Cleanup(server.Close)

	request, err := http.NewRequest(http.MethodGet, server.URL+"/system/namespaces", nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("want unauthenticated status %d, got %d", http.StatusUnauthorized, response.StatusCode)
	}

	request, err = http.NewRequest(http.MethodGet, server.URL+"/system/namespaces", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.SetBasicAuth(user, password)
	response, err = server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("want authenticated status %d, got %d", http.StatusNotImplemented, response.StatusCode)
	}
}
