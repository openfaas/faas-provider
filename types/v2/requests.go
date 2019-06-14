// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v2

// Secret for underlying orchestrator
type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

// ScaleServiceRequest represents a scale command
type ScaleServiceRequest struct {
	ServiceName string `json:"serviceName"`
	Replicas    uint64 `json:"replicas"`
}

// InfoRequest provides information about the underlying provider
type InfoRequest struct {
	Provider      string          `json:"provider"`
	Version       ProviderVersion `json:"version"`
	Orchestration string          `json:"orchestration"`
}

// ProviderVersion provides the commit sha and release version number of the underlying provider
type ProviderVersion struct {
	SHA     string `json:"sha"`
	Release string `json:"release"`
}
