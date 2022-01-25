package sdk

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSdk_GetNamespaces_TwoNamespaces(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		rw.Write([]byte(`["openfaas-fn","dev"]`))
	}))

	sU, _ := url.Parse(s.URL)

	sdk := NewSDK(sU, nil, http.DefaultClient)
	ns, err := sdk.GetNamespaces()
	if err != nil {
		t.Fatalf("wanted no error, but got: %s", err)
	}
	want := 2
	if len(ns) != want {
		t.Fatalf("want %d namespaces, got: %d", want, len(ns))
	}
	wantNS := []string{"openfaas-fn", "dev"}
	gotNS := 0

	for _, n := range ns {
		for _, w := range wantNS {
			if n == w {
				gotNS++
			}
		}
	}
	if gotNS != len(wantNS) {
		t.Fatalf("want %d namespaces, got: %d", len(wantNS), gotNS)
	}
}

func TestSdk_GetNamespaces_NoNamespaces(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		rw.Write([]byte(`[]`))
	}))

	sU, _ := url.Parse(s.URL)

	sdk := NewSDK(sU, nil, http.DefaultClient)
	ns, err := sdk.GetNamespaces()
	if err != nil {
		t.Fatalf("wanted no error, but got: %s", err)
	}
	want := 0
	if len(ns) != want {
		t.Fatalf("want %d namespaces, got: %d", want, len(ns))
	}
	wantNS := []string{}
	gotNS := 0

	for _, n := range ns {
		for _, w := range wantNS {
			if n == w {
				gotNS++
			}
		}
	}
	if gotNS != len(wantNS) {
		t.Fatalf("want %d namespaces, got: %d", len(wantNS), gotNS)
	}
}
