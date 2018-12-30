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

### Features

[ ] Plugin support
[ ] Metrics
[ ] Background thread scalability (support exponential back-off pressure per resource)

### Resources

#### Google Cloud Platform

[ ] VPC Firewall Rules
[ ] Cloud DNS Records
[ ] BigTable Instances
[ ] Datastore & Firestore Instances
[ ] Cloud SQL Instances
[ ] Spanner Instances
[ ] Memorystore Instances
[ ] Filestore Instances
[ ] BigQuery datasets
[ ] Pub/Sub Topics & Subscriptions

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
