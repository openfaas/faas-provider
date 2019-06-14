// Copyright 2019 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package v2

// Function describes an OpenFaaS function
type Function struct {
	// Service is the name of the function
	Service string `json:"service"`

	// ServiceAccount represents the identity of the function for the orchestration layer
	ServiceAccount string `json:"serviceAccount"`

	// Image corresponds to a container image
	Image string `json:"image"`

	// Network is specific to Docker Swarm - default overlay network is: func_functions
	Network string `json:"network"`

	// EnvProcess corresponds to the fprocess variable for your container watchdog.
	EnvProcess string `json:"envProcess"`

	// EnvVars provides overrides for functions.
	EnvVars map[string]string `json:"envVars"`

	// RegistryAuth is the registry authentication (optional)
	// in the same encoded format as Docker native credentials
	// (see ~/.docker/config.json)
	RegistryAuth string `json:"registryAuth,omitempty"`

	// Constraints are specific to back-end orchestration platform
	Constraints []string `json:"constraints"`

	// Secrets list of secrets to be made available to function
	Secrets []string `json:"secrets"`

	// Labels are metadata for functions which may be used by the
	// back-end for making scheduling or routing decisions
	Labels *map[string]string `json:"labels"`

	// Annotations are metadata for functions which may be used by the
	// back-end for management, orchestration, events and build tasks
	Annotations *map[string]string `json:"annotations"`

	// Limits for function
	Limits *FunctionResources `json:"limits"`

	// Requests of resources requested by function
	Requests *FunctionResources `json:"requests"`

	// ReadOnlyRootFilesystem removes write-access from the root filesystem
	// mount-point.
	ReadOnlyRootFilesystem bool `json:"readOnlyRootFilesystem"`

	// FunctionHealthCheck represents the custom HTTP health-check path and check initial delay
	HealthCheck *FunctionHealthCheck `json:"healthCheck"`

	// Scaling represents the minimum (initial), maximum replica count and scaling factor
	Scaling *FunctionScaling `json:"scaling"`

	// FunctionTopic defines the MQ topics that this function is subscribed to
	Topics []FunctionTopic `json:"topics"`
}

// FunctionResources represents CPU and memory resources for an OpenFaaS function
type FunctionResources struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

// FunctionHealthCheck represents the custom HTTP health-check path and check initial delay
type FunctionHealthCheck struct {
	Path         string `json:"path"`
	InitialDelay string `json:"initialDelay"`
}

// FunctionScaling represents the minimum (initial), maximum replica count and scaling factor
type FunctionScaling struct {
	// Min defaults to one replica
	Min *int `json:"min,omitempty"`
	// Max defaults to 20 replicas
	Max *int `json:"max,omitempty"`
	// Factor defaults to 20%
	// Has to be a value between 0-100 (including borders)
	// Setting the factory to zero disables the auto scaling
	Factor *int `json:"factor,omitempty"`
	// ToZero enables a function to be scaled to zero
	ToZero bool `json:"zero"`
}

// FunctionTopic represents an OpenFaaS MQ topic name
type FunctionTopic string
