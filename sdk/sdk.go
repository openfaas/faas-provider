package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/openfaas/faas-provider/auth"
	"github.com/openfaas/faas-provider/types"
)

// SDK is an SDK for managing OpenFaaS functions
type SDK struct {
	GatewayURL  *url.URL
	Client      *http.Client
	Credentials *auth.BasicAuthCredentials
}

// NewSDK creates an SDK for managing OpenFaaS
func NewSDK(gatewayURL *url.URL, credentials *auth.BasicAuthCredentials, client *http.Client) *SDK {
	return &SDK{
		GatewayURL:  gatewayURL,
		Client:      http.DefaultClient,
		Credentials: credentials,
	}
}

// GetNamespaces get openfaas namespaces
func (s *SDK) GetNamespaces() ([]string, error) {
	u := s.GatewayURL
	namespaces := []string{}
	u.Path = "/system/namespaces"

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return namespaces, fmt.Errorf("unable to create request: %s, error: %w", u.String(), err)
	}

	if s.Credentials != nil {
		req.SetBasicAuth(s.Credentials.User, s.Credentials.Password)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return namespaces, fmt.Errorf("unable to make request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bytesOut, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return namespaces, err
	}

	if res.StatusCode == http.StatusUnauthorized {
		return namespaces, fmt.Errorf("check authorization, status code: %d", res.StatusCode)
	}

	if len(bytesOut) == 0 {
		return namespaces, nil
	}

	if err := json.Unmarshal(bytesOut, &namespaces); err != nil {
		return namespaces, fmt.Errorf("unable to marshal to JSON: %s, error: %w", string(bytesOut), err)
	}

	return namespaces, err
}

// GetFunctions lists all functions
func (s *SDK) GetFunctions(namespace string) ([]types.FunctionStatus, error) {
	u := s.GatewayURL

	u.Path = "/system/functions"

	if len(namespace) > 0 {
		query := u.Query()
		query.Set("namespace", namespace)
		u.RawQuery = query.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return []types.FunctionStatus{}, fmt.Errorf("unable to create request for %s, error: %w", u.String(), err)
	}

	if s.Credentials != nil {
		req.SetBasicAuth(s.Credentials.User, s.Credentials.Password)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return []types.FunctionStatus{}, fmt.Errorf("unable to make HTTP request: %w", err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, _ := ioutil.ReadAll(res.Body)

	functions := []types.FunctionStatus{}
	if err := json.Unmarshal(body, &functions); err != nil {
		return []types.FunctionStatus{},
			fmt.Errorf("unable to unmarshal value: %q, error: %w", string(body), err)
	}

	return functions, nil
}

// ScaleFunction scales a function to a number of replicas
func (s *SDK) ScaleFunction(ctx context.Context, functionName, namespace string, replicas uint64) error {

	scaleReq := types.ScaleServiceRequest{
		ServiceName: functionName,
		Replicas:    replicas,
	}

	var err error

	bodyBytes, _ := json.Marshal(scaleReq)
	bodyReader := bytes.NewReader(bodyBytes)

	u := s.GatewayURL

	functionPath := filepath.Join("/system/scale-function", functionName)
	if len(namespace) > 0 {
		query := u.Query()
		query.Set("namespace", namespace)
		u.RawQuery = query.Encode()
	}

	u.Path = functionPath

	req, err := http.NewRequest(http.MethodPost, u.String(), bodyReader)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", u.String(), err)
	}

	if s.Credentials != nil {
		req.SetBasicAuth(s.Credentials.User, s.Credentials.Password)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot connect to OpenFaaS on URL: %s, error: %s", s.GatewayURL, err)

	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusAccepted, http.StatusOK, http.StatusCreated:
		break

	case http.StatusNotFound:
		return fmt.Errorf("function %s not found", functionName)

	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized action, please setup authentication for this server")

	default:
		var err error
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("server returned unexpected status code %d, message: %q", res.StatusCode, string(bytesOut))
	}
	return nil
}
