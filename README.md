faas-provider
==============

This faas-provider can be used to write your own back-end for OpenFaaS. The Golang SDK can be vendored into your project so that you can provide a provider which is compliant and compatible with the OpenFaaS gateway.

![Conceptual diagram](docs/conceptual.png)

The faas-provider provides CRUD for functions and an invoke capability. If you complete the required endpoints then you will be able to use your container orchestrator or back-end system with the existing OpenFaaS ecosystem and tooling.

> See also: [backends guide](https://github.com/openfaas/faas/blob/master/guide/deprecated/backends.md)

### Recommendations

The following is used in OpenFaaS and recommended for those seeking to build their own back-ends:

* License: MIT
* Language: Golang 

### How to use this project

All the required HTTP routes are configured automatically including a HTTP server on port 8080. Your task is to implement the supplied HTTP handler functions.

For an example see the [server.go](https://github.com/openfaas/faas-netes/blob/master/server.go) file in the [faas-netes](https://github.com/openfaas/faas-netes) Kubernetes backend.

I.e.:

```go
import (
	bootTypes "github.com/openfaas/faas-provider/types/v1"
	bootstrap "github.com/openfaas/faas-provider/bootstrap/v1"
)

bootstrapHandlers := bootTypes.FaaSHandlers{
    FunctionProxy:  handlers.MakeProxy(),
    DeleteHandler:  handlers.MakeDeleteHandler(clientset),
    DeployHandler:  handlers.MakeDeployHandler(clientset),
    FunctionReader: handlers.MakeFunctionReader(clientset),
    ReplicaReader:  handlers.MakeReplicaReader(clientset),
    ReplicaUpdater: handlers.MakeReplicaUpdater(clientset),
    InfoHandler:    handlers.MakeInfoHandler(),
}

var port int
port = 8080
bootstrapConfig := bootTypes.FaaSConfig{
    ReadTimeout:  time.Second * 8,
    WriteTimeout: time.Second * 8,
    TCPPort:      &port,
}

bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
```

### Upgrade to OpenFaaS provider v2 packages

Example of v2 provider bootstrap:

```go
import bootstrap "github.com/openfaas/faas-provider/bootstrap/v2"

env := bootstrap.EnvReader{}

config := bootstrap.ApiServerConfig{
	ReadTimeout:  env.ParseDuration("read_timeout", time.Second*10),
	WriteTimeout: env.ParseDuration("write_timeout", time.Second*10),
	TCPPort:      env.ParseInt("port", 8080),
	EnableHealth: true,
}

handlers = bootstrap.ApiServerHandlers{
	FunctionProxy:  handlers.MakeProxy(namespace, config.ReadTimeout),
	DeleteHandler:  handlers.MakeDeleteHandler(namespace, clientset),
	DeployHandler:  handlers.MakeDeployHandler(namespace, factory),
	FunctionReader: handlers.MakeFunctionReader(namespace, clientset),
	ReplicaReader:  handlers.MakeReplicaReader(namespace, clientset),
	ReplicaUpdater: handlers.MakeReplicaUpdater(namespace, clientset),
	UpdateHandler:  handlers.MakeUpdateHandler(namespace, factory),
	HealthHandler:  handlers.MakeHealthHandler(),
	InfoHandler:    handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommit),
	SecretHandler:  handlers.MakeSecretHandler(namespace, clientset),
}

srv := bootstrap.NewApiServer(config, handlers)

log.Fatal(srv.ListenAndServe())
```

### Need help?

Join `#faas-provider` on [OpenFaaS Slack](https://docs.openfaas.com/community/)
