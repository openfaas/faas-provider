package proxy

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/openfaas/faas/gateway/requests"
)

const (
	watchdogPort       = 8080
	defaultContentType = "text/plain"
)

// NewHandlerFunc creates a standard http HandlerFunc to proxy function requests to the function.
// The FaaS provider implementation is responsible for providing the resolver function implementation.
// resolver will receive the function name and should return the address of the function service.
func NewHandlerFunc(timeout time.Duration, resolver func(string) string) http.HandlerFunc {
	proxyClient := http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   timeout,
				KeepAlive: 1 * time.Second,
			}).DialContext,
			IdleConnTimeout:       120 * time.Millisecond,
			ExpectContinueTimeout: 1500 * time.Millisecond,
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			defer r.Body.Close()
		}

		switch r.Method {
		case http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodGet:

			proxyRequest(w, r, proxyClient, resolver)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// proxyRequest handles the actual resolution of and then request to the function service.
func proxyRequest(w http.ResponseWriter, originalReq *http.Request, proxyClient http.Client, resolver func(string) string) {

	pathVars := mux.Vars(originalReq)
	// extraPath := pathVars["params"]
	functionName := pathVars["name"]
	if functionName == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("provide v valid route /function/function_name."))
		return
	}

	functionAddr := resolver(functionName)
	if functionAddr == "" {
		// TODO: Should record the 404/not found error in Prometheus.
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("Cannot find service: %s.", functionName)))
		return
	}

	defer func(when time.Time) {
		seconds := time.Since(when).Seconds()
		log.Printf("%s took %f seconds\n", functionName, seconds)
	}(time.Now())

	forwardReq := requests.NewForwardRequest(originalReq.Method, *originalReq.URL)
	url := forwardReq.ToURL(functionAddr, watchdogPort)
	proxyReq, _ := http.NewRequest(originalReq.Method, url, originalReq.Body)
	defer proxyReq.Body.Close()

	copyHeaders(&proxyReq.Header, &originalReq.Header)

	response, err := proxyClient.Do(proxyReq)
	if err != nil {
		log.Println("[ERROR]", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		buf := bytes.NewBufferString("Can't reach service for: " + functionName)
		w.Write(buf.Bytes())
		return
	}

	clientHeader := w.Header()
	copyHeaders(&clientHeader, &response.Header)
	w.Header().Set("Content-Type", getContentType(response.Header, originalReq.Header))

	w.WriteHeader(http.StatusOK)
	io.Copy(w, response.Body)
}

// copyHeaders clones the header values from the source into the destination.
func copyHeaders(destination *http.Header, source *http.Header) {
	for k, v := range *source {
		vClone := make([]string, len(v))
		copy(vClone, v)
		(*destination)[k] = vClone
	}
}

// getContentType resolves the correct Content-Tyoe for a proxied function.
func getContentType(request http.Header, proxyResponse http.Header) (headerContentType string) {
	responseHeader := proxyResponse.Get("Content-Type")
	requestHeader := request.Get("Content-Type")

	if len(responseHeader) > 0 {
		headerContentType = responseHeader
	} else if len(requestHeader) > 0 {
		headerContentType = requestHeader
	} else {
		headerContentType = defaultContentType
	}

	return headerContentType
}
