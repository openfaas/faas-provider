// Copyright (c) OpenFaaS Author(s). All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package types

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"strings"
)

// BasicAuthCredentials are the user and password strings
type BasicAuthCredentials struct {
	User     string
	Password string
}

// ReadBasicAuth interface for reading the credentials
type ReadBasicAuth interface {
	Read() (error, *BasicAuthCredentials)
}

// ReadBasicAuthFromDisk defines the path for the secret
type ReadBasicAuthFromDisk struct {
	SecretMountPath string
}

func (r *ReadBasicAuthFromDisk) Read() (*BasicAuthCredentials, error) {
	var credentials *BasicAuthCredentials

	if len(r.SecretMountPath) == 0 {
		return nil, fmt.Errorf("invalid SecretMountPath specified for reading secrets")
	}

	userPath := path.Join(r.SecretMountPath, "basic-auth-user")
	user, userErr := ioutil.ReadFile(userPath)
	if userErr != nil {
		return nil, fmt.Errorf("unable to load %s", userPath)
	}

	userPassword := path.Join(r.SecretMountPath, "basic-auth-password")
	password, passErr := ioutil.ReadFile(userPassword)
	if passErr != nil {
		return nil, fmt.Errorf("Unable to load %s", userPassword)
	}

	re := regexp.MustCompile(`\r?\n`)
	user = re.ReplaceAll(user, []byte(" "))
	password = re.ReplaceAll(password, []byte(" "))

	credentials = &BasicAuthCredentials{
		User:     strings.TrimSpace(string(user)),
		Password: strings.TrimSpace(string(password)),
	}

	return credentials, nil
}
