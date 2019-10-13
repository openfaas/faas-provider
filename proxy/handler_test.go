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

	config := types.FaaSConfig{ReadTimeout: time.Second}
	NewHandlerFunc(config, nil)
}

func Test_NewHandlerFunc_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not panic if resolver is not nil")
		}
	}()

	config := types.FaaSConfig{ReadTimeout: time.Second}
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{})
	if proxyFunc == nil {
		t.Errorf("proxy handler func is nil")
	}
}

func Test_ProxyHandler_NonAllowedMethods(t *testing.T) {
	config := types.FaaSConfig{ReadTimeout: time.Second}
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{})

	nonAllowedMethods := []string{
		http.MethodHead, http.MethodConnect, http.MethodOptions, http.MethodTrace,
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
	config := types.FaaSConfig{ReadTimeout: time.Second}
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{"", nil})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req = mux.SetURLVars(req, map[string]string{"name": ""})

	proxyFunc(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status code `%d`, got `%d`", http.StatusBadRequest, w.Code)
	}

	respBody := strings.TrimSpace(w.Body.String())
	if respBody != errMissingFunctionName {
		t.Errorf("expected error message `%s`, got `%s`", errMissingFunctionName, respBody)
	}
}

func Test_ProxyHandler_ResolveError(t *testing.T) {
	logs := &bytes.Buffer{}
	log.SetOutput(logs)

	resolveErr := errors.New("can not find test service `foo`")

	config := types.FaaSConfig{ReadTimeout: time.Second}
	proxyFunc := NewHandlerFunc(config, &testBaseURLResolver{"", resolveErr})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req = mux.SetURLVars(req, map[string]string{"name": "foo"})

	proxyFunc(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status code `%d`, got `%d`", http.StatusBadRequest, w.Code)
	}

	respBody := strings.TrimSpace(w.Body.String())
	if respBody != "Cannot find service: foo." {
		t.Errorf("expected error message `%s`, got `%s`", "Cannot find service: foo.", respBody)
	}

	if !strings.Contains(logs.String(), resolveErr.Error()) {
		t.Errorf("expected logs to contain `%s`", resolveErr.Error())
	}
}

func Test_ProxyHandler_Proxy_Success(t *testing.T) {
	t.Skip("Test not implemented yet")
	// testFuncService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	w.WriteHeader(http.StatusOK)
	// }))
	// proxyFunc := NewHandlerFunc(time.Second, &testBaseURLResolver{testFuncService.URL, nil})

	// w := httptest.NewRecorder()
	// req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	// req = mux.SetURLVars(req, map[string]string{"name": "foo"})

	// proxyFunc(w, req)
}
