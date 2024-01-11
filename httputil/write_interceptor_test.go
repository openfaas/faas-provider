package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_WriteCountsBytes(t *testing.T) {
	w := httptest.NewRecorder()
	wi := NewHttpWriteInterceptor(w)

	writeStr := "hello world"
	wi.Write([]byte(writeStr))

	want := int64(len(writeStr))
	got := wi.BytesWritten()
	if got != want {
		t.Errorf("want bytes: %d, got %d", want, got)
	}
}

func Test_WriteGetsStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	wi := NewHttpWriteInterceptor(w)

	wi.WriteHeader(http.StatusTeapot)

	want := http.StatusTeapot
	got := wi.Status()
	if got != want {
		t.Errorf("want status code: %d, got %d", want, got)
	}
}

func Test_WriteGetsStatusCode_WithoutWriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	wi := NewHttpWriteInterceptor(w)

	wi.Write([]byte("hello world"))

	want := http.StatusOK
	got := wi.Status()
	if got != want {
		t.Errorf("want default status code: %d, got %d", want, got)
	}
}
