package proxy

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas-provider/types"
)

type testBaseURLResolver struct {
	testServerBase string
	err            error
}

func (tr *testBaseURLResolver) Resolve(name string) (url.URL, error) {
	if tr.err != nil {
		return url.URL{}, tr.err
	}

	return url.URL{
		Scheme: "http",
		Host:   tr.testServerBase,
	}, nil
}

func Test_NewHandlerFunc_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("should panic if resolver is nil")
		}
	}()

	config := types.FaaSConfig{ReadTimeout: 100 * time.Millisecond}
	NewHandlerFunc(config, nil, false)
}

func Test_NewHandlerFunc_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not panic if resolver is not nil")
		}
	}()

	config := types.FaaSConfig{ReadTimeout: 100 * time.Millisecond}
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{}, false)
	if proxyFunc == nil {
		t.Errorf("proxy handler func is nil")
	}
}

func Test_ProxyHandler_NonAllowedMethods(t *testing.T) {
	config := types.FaaSConfig{ReadTimeout: 100 * time.Millisecond}
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{}, false)

	nonAllowedMethods := []string{
		http.MethodConnect, http.MethodTrace,
	}

	for _, method := range nonAllowedMethods {
		t.Run(method+" method is not allowed", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(method, "http://example.com/foo", nil)
			proxyFunc(w, req)
			resp := w.Result()
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("expected status code `%d`, got `%d`", http.StatusMethodNotAllowed, resp.StatusCode)
			}
		})
	}
}

func Test_ProxyHandler_MissingFunctionNameError(t *testing.T) {
	config := types.FaaSConfig{ReadTimeout: 100 * time.Millisecond}
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{"", nil}, false)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req = mux.SetURLVars(req, map[string]string{"name": ""})

	proxyFunc(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status code `%d`, got `%d`", http.StatusBadRequest, w.Code)
	}
	want := "Provide function name in the request path"
	respBody := strings.TrimSpace(w.Body.String())
	if respBody != want {
		t.Errorf("want error message %q, but got %q", want, respBody)
	}
}

func Test_ProxyHandler_ResolveError(t *testing.T) {
	logs := &bytes.Buffer{}
	log.SetOutput(logs)

	resolveErr := errors.New("can not find test service `foo`")

	config := types.FaaSConfig{ReadTimeout: 100 * time.Millisecond}
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{"", resolveErr}, false)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "foo"})

	proxyFunc(w, req)

	wantStatus := http.StatusServiceUnavailable

	if w.Code != wantStatus {
		t.Errorf("status code want `%d`, but got `%d`", wantStatus, w.Code)
	}

	want := `No endpoints available for: foo.`
	respBody := strings.TrimSpace(w.Body.String())
	if respBody != want {
		t.Errorf("want error `%s`, but got `%s`", want, respBody)
	}

	if !strings.Contains(logs.String(), resolveErr.Error()) {
		t.Errorf("expected logs to contain `%s`", resolveErr.Error())
	}

	internalErrorHeader := w.Header().Get("X-OpenFaaS-Internal")
	wantHeaderValue := "proxy"
	if internalErrorHeader != wantHeaderValue {
		t.Errorf("expected X-OpenFaaS-Internal header to be %s, got %s", wantHeaderValue, internalErrorHeader)
	}
}

func Test_ProxyHandler_Proxy_Success(t *testing.T) {
	testFuncService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	config := types.FaaSConfig{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	serverURL := strings.TrimPrefix(testFuncService.URL, "http://")
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{serverURL, nil}, false)

	nonAllowedMethods := []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodGet,
		http.MethodOptions,
		http.MethodHead,
	}
	for _, method := range nonAllowedMethods {
		t.Run(method+" method is allowed", func(t *testing.T) {
			w := httptest.NewRecorder()

			req := httptest.NewRequest(method, "http://example.com/foo", nil)
			req = mux.SetURLVars(req, map[string]string{"name": "foo"})

			proxyFunc(w, req)
			resp := w.Result()
			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("expected status code `%d`, got `%d`", http.StatusNoContent, resp.StatusCode)
			}

			if v := resp.Header.Get("X-OpenFaaS-Internal"); v != "" {
				t.Fatalf("expected X-OpenFaaS-Internal header to be empty, got %q", v)
			}

		})
	}
}

func Test_ProxyHandler_Proxy_FailsMidFlight(t *testing.T) {
	var svr *httptest.Server

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		svr.Close()
		// w.WriteHeader(http.StatusOK)
	})
	svr = httptest.NewServer(testHandler)

	config := types.FaaSConfig{
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	}

	serverURL := strings.TrimPrefix(svr.URL, "http://")
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{serverURL, nil}, false)

	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "http://example.com/foo", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "foo"})

	proxyFunc(w, req)
	resp := w.Result()
	wantCode := http.StatusInternalServerError
	if resp.StatusCode != wantCode {
		t.Fatalf("want status code `%d`, got `%d`", wantCode, resp.StatusCode)
	}

	if v := resp.Header.Get("X-OpenFaaS-Internal"); v != "proxy" {
		t.Errorf("expected X-OpenFaaS-Internal header to be `proxy`, got %s", v)
	}
}
