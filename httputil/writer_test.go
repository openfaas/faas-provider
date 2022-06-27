package httputil

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_wn(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := NewHttpWriteInterceptor(w)
		cn := ww.ResponseWriter.(http.CloseNotifier)

		fmt.Println("CloseNotfier?", cn)
		writeAccepted(ww, r)

		fmt.Println("Closed")
		fmt.Println(ww.StatusCode)
	}))

	defer s.Close()

	req, _ := http.NewRequest(http.MethodGet, s.URL, nil)
	res, err := http.DefaultClient.Do(req)

	if err != nil {
		t.Fatal(err)
	}

	log.Println(res.Status)
}

func writeAccepted(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusAccepted)
}
