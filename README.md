# gitzup

[![Go Report Card](https://goreportcard.com/badge/github.com/kfirz/gitzup?style=flat-square)](https://goreportcard.com/report/github.com/kfirz/gitzup)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/kfirz/gitzup)
[![Release](https://img.shields.io/github/release/kfirz/gitzup.svg?style=flat-square)](https://github.com/kfirz/gitzup/releases/latest)

Gitzup aims to be an opinionated DevOps platform for managing all development aspects, from infrastructure, through coding, to deployment.

## Components

Gitzup is composed of the following components:

* Kubernetes custom resource definitions (CRDs) for external resources. This component harnesses Kubernetes's asynchronous nature and resilience to manage deployment for external resources outside of the cluster, such as DNS zones, reserved IP addresses, etc.

## Status

Gitzup is in an early stage in its development, and not yet ready for consumption. Currently, only a single component is partially implemented (custom CRD-based deployment of cloud provider resources).

## Installation

TBD.
 
## Roadmap / TODOs

[ ] Plugin support for custom resources
[ ] Metrics & alerts
[ ] Background thread interval should be one minute
[ ] Background thread should support back-off pressure

## Development

#### Prerequisites

* Working [Minikube](https://kubernetes.io/docs/tasks/tools/install-kubectl/) cluster

* GNU Make

* Working Go workspace


#### Deploying to Minikube

```bash
$ make minikube
```

## Contributing

Please see our [contributing](./CONTRIBUTING.md) document.
