package httputil

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFirstWriteSetsStatusCode(t *testing.T) {

	w := httptest.NewRecorder()
	ww := NewHttpWriteInterceptor(w)
	ww.Write([]byte("{"))
	ww.Write([]byte(`"value": "ok"}`))

	if ww.statusCode != http.StatusOK {
		t.Fatalf("incorrect status code: %d", ww.statusCode)
	}

	if w.Result().StatusCode != ww.statusCode {
		t.Fatalf("incorrect status code in the original response object: %d", w.Result().StatusCode)
	}

	out, _ := ioutil.ReadAll(w.Result().Body)
	if string(out) != `{"value": "ok"}` {
		t.Fatalf("incorrect response content: %q", out)
	}

}

func Test_wn(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := NewHttpWriteInterceptor(w)
		cn := ww.ResponseWriter.(http.CloseNotifier)

		fmt.Println("CloseNotfier?", cn)
		writeAccepted(ww, r)

		fmt.Println("Closed")
		fmt.Println(ww.statusCode)
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
