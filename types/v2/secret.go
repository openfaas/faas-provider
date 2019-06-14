// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v2

// Secret for underlying orchestrator
type Secret struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}
