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

func TestDecorateWithBasicAuth(t *testing.T) {
	credentials := &auth.BasicAuthCredentials{
		User:     "admin",
		Password: "correct-horse-battery-staple",
	}

	tests := []struct {
		name    string
		method  string
		path    string
		status  int
		handler func(*types.FaaSHandlers) http.HandlerFunc
	}{
		{"list namespaces", http.MethodGet, "/system/namespaces", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.ListNamespaces }},
		{"get namespace", http.MethodGet, "/system/namespace/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.MutateNamespace }},
		{"create namespace", http.MethodPost, "/system/namespace/test", http.StatusCreated, func(h *types.FaaSHandlers) http.HandlerFunc { return h.MutateNamespace }},
		{"update namespace", http.MethodPut, "/system/namespace/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.MutateNamespace }},
		{"delete namespace", http.MethodDelete, "/system/namespace/test", http.StatusAccepted, func(h *types.FaaSHandlers) http.HandlerFunc { return h.MutateNamespace }},
		{"FunctionLister", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.FunctionLister }},
		{"DeployFunction", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.DeployFunction }},
		{"DeleteFunction", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.DeleteFunction }},
		{"UpdateFunction", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.UpdateFunction }},
		{"FunctionStatus", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.FunctionStatus }},
		{"ScaleFunction", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.ScaleFunction }},
		{"Info", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.Info }},
		{"Secrets", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.Secrets }},
		{"Logs", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.Logs }},
		{"Telemetry", http.MethodGet, "/system/test", http.StatusOK, func(h *types.FaaSHandlers) http.HandlerFunc { return h.Telemetry }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := 0
			stub := func(w http.ResponseWriter, _ *http.Request) {
				called++
				w.WriteHeader(tc.status)
			}

			handlers := types.FaaSHandlers{
				ListNamespaces:  stub,
				MutateNamespace: stub,
				FunctionLister:  stub,
				DeployFunction:  stub,
				DeleteFunction:  stub,
				UpdateFunction:  stub,
				FunctionStatus:  stub,
				ScaleFunction:   stub,
				Info:            stub,
				Secrets:         stub,
				Logs:            stub,
				Telemetry:       stub,
			}
			decorateWithBasicAuth(&handlers, credentials)

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

			if authenticatedResponse.Code != tc.status {
				t.Fatalf("want authenticated status %d, got %d", tc.status, authenticatedResponse.Code)
			}
			if called != 1 {
				t.Fatalf("want protected handler to be called once, got %d calls", called)
			}
		})
	}
}

func TestDecorateWithBasicAuthAllowsNilHandlers(t *testing.T) {
	stub := func(http.ResponseWriter, *http.Request) {}
	for _, handlers := range []types.FaaSHandlers{
		{},
		{ListNamespaces: stub},
		{MutateNamespace: stub},
	} {
		listNil := handlers.ListNamespaces == nil
		mutateNil := handlers.MutateNamespace == nil
		decorateWithBasicAuth(&handlers, &auth.BasicAuthCredentials{})
		if (handlers.ListNamespaces == nil) != listNil {
			t.Fatal("namespace lister availability changed")
		}
		if (handlers.MutateNamespace == nil) != mutateNil {
			t.Fatal("namespace mutation availability changed")
		}
	}
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
