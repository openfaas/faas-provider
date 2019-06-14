// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v2

import (
	"time"
)

// ApiServerConfig represents the configuration of the OpenFaaS API server
type ApiServerConfig struct {
	TCPPort         *int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	EnableHealth    bool
	EnableBasicAuth bool
	SecretMountPath string
}
