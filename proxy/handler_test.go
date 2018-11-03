package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type testBaseURLResolver struct {
	testServerBase string
}

func (tr *testBaseURLResolver) Resolve(name string) (url.URL, error) {
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

	NewHandlerFunc(time.Second, nil)
}

func Test_NewHandlerFunc_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("should not panic if resolver is not nil")
		}
	}()

	proxyFunc := NewHandlerFunc(time.Second, &testBaseURLResolver{})
	if proxyFunc == nil {
		t.Errorf("proxy handler func is nil")
	}
}

func Test_ProxyHandler_NonAllowedMethods(t *testing.T) {

	proxyFunc := NewHandlerFunc(time.Second, &testBaseURLResolver{})

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
