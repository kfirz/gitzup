# Contributing

When contributing to this repository, please first discuss the change you wish to make via issue, email, or any other method with the owners of this repository before making a change. This usually saves time & effort.

Please note we have a [code of conduct](./CODE_OF_CONDUCT.md), please follow it in all your interactions with the project.

## Pull Request process

Aside from the actual change in source code, please ensure your PR update any relevant tests and/or adds new tests as necessary. PRs that lower test coverage, or cause test failures, will not be accepted.

For cases where the change affects information displayed in the documentation, please ensure the PR updates the documentation as well (eg. `README.md`).

## Development

### Prerequisites

* **Minikube**
  
  This is crucial for developing and testing your changes. Once you [install Minikube](https://kubernetes.io/docs/tasks/tools/install-kubectl/), this workflow (roughly) should work for you:
  
  (note that the `--vm-driver=hyperkit` is only needed on Mac, not supported on Linux/Windows)
  
  ```bash
  $ minikube version
  minikube version: v0.32.0
  
  $ minikube start --kubernetes-version v1.13.1 --vm-driver=hyperkit
  Starting local Kubernetes v1.13.1 cluster...
  Starting VM...
  Getting VM IP address...
  Moving files into cluster...
  Setting up certs...
  Connecting to cluster...
  Setting up kubeconfig...
  Stopping extra container runtimes...
  Starting cluster components...
  Verifying kubelet health ...
  Verifying apiserver health ...Kubectl is now configured to use the cluster.
  Loading cached images from config file.
  
  
  Everything looks great. Please enjoy minikube!

  ```

* **GNU Make**

  This is required as Gitzup is using `Make` as the build tool. Note that strictly speaking, `Make` is not _really_ required, just makes life easier.
  
  Note that `Make` is available by default on Mac and easily installable on Linux.

* **Working Go workspace**

  Since Gitzup is written in Go, you should be familiar and comfortable with setting up a Go workspace, and how to do that is outside the scope of this manual.

### Building

Once you clone the repository, make sure you fetch all the project's dependencies:

```bash
make dep
```

You can now build the project either via an IDE or via the command line, as so:

```bash
$ make manager
```

To test:

```bash
$ make test
```

### Minikube Development

We recommend the following development workflow:

1. Set up a Minikube instance

2. Deploy the CRDs 

   ```bash
   $ make deploy-crds
   ```

3. Run the Gitzup main from your IDE (eg. [Goland](https://www.jetbrains.com/go/))
