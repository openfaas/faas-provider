package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_StatusIsRecorded(t *testing.T) {
	wantCode := http.StatusAccepted
	gotCode := 0

	next := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(wantCode)

	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := NewHttpWriteInterceptor(w)
		next(ww, r)
		gotCode = ww.Status()
	}))

	defer func() {
		s.Close()
	}()

	req, _ := http.NewRequest("GET", s.URL, nil)
	_, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatalf("Error doing request: %v", err)
	}

	if gotCode != wantCode {
		t.Errorf("got code %d, want %d", gotCode, wantCode)
	}
}
