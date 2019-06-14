// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v2

import (
	"os"
	"strconv"
	"time"
)

// EnvReader wraps os.Getenv and allows fallback values for not found environment variables
type EnvReader struct {
}

func (*EnvReader) ParseString(key string, fallback string) string {
	val := os.Getenv(key)
	if len(val) > 0 {
		return val
	}
	return fallback
}

func (*EnvReader) ParseInt(key string, fallback int) int {
	val := os.Getenv(key)
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return parsedVal
		}
	}
	return fallback
}

func (*EnvReader) ParseDuration(key string, fallback time.Duration) time.Duration {
	val := os.Getenv(key)
	if len(val) > 0 {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal >= 0 {
			return time.Duration(parsedVal) * time.Second
		}
	}

	duration, durationErr := time.ParseDuration(val)
	if durationErr != nil {
		return fallback
	}

	return duration
}

func (*EnvReader) ParseBool(key string, fallback bool) bool {
	val := os.Getenv(key)
	switch val {
	case "1", "true":
		return true
	case "0", "false":
		return false
	default:
		return fallback
	}
}
