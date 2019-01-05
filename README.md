# gitzup

[![Go Report Card](https://goreportcard.com/badge/github.com/kfirz/gitzup?style=flat-square)](https://goreportcard.com/report/github.com/kfirz/gitzup)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/kfirz/gitzup)
[![Release](https://img.shields.io/github/release/kfirz/gitzup.svg?style=flat-square)](https://github.com/kfirz/gitzup/releases/latest)

Gitzup aims to implement & automate continuous infrastructure-as-code by utilizing Kubernetes' asynchronous nature for ensuring the actual infrastructure state matches the desired infrastructure state.

The key is Kubernetes custom resource definitions (CRD). Gitzup implements CRDs that act as proxies to actual cloud resources _external to the cluster_, and continually monitors & adjusts the external resources to match the deployed CRDs in the cluster.
 
## Status

![ALPHA SOFTWARE](https://img.shields.io/badge/status:-ALPHA-yellow.svg)

Gitzup is in an early stage in its development, and is not yet ready for consumption.

## Installation

1. Create a cluster (outside the scope of this document).

2. Ensure your `kubectl` is configured to access this cluster.

3. If you're running in Google Cloud Platform (GCP):

   ```bash
   $ PROJECT=my-project
   $ gcloud iam service-accounts create gitzup --display-name "Gitzup"
   $ gcloud projects add-iam-policy-binding ${PROJECT} \
         --member serviceAccount:gitzup@${PROJECT}.iam.gserviceaccount.com \
         --role roles/editor
   ```

   You might want to customize the role(s) you want to give the service account, according to the type of resources you plan to use Gitzup to deploy for you.

4. Run the following command (make sure the update `GITZUP_RELEASE` accordingly):

   ```bash
   $ GITZUP_RELEASE=x.y.z
   $ kubectl apply -f https://github.com/kfirz/gitzup/releases/download/${GITZUP_RELEASE}/gitzup.yaml  
   ```

## Usage

TBD.
         
## Contributing

Please read the [contributing](./CONTRIBUTING.md) guide.

## Copyright and license

Copyright the authors and contributors. See individual source files for details.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright
    notice, this list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright
    notice, this list of conditions and the following disclaimer in the
    documentation and/or other materials provided with the distribution.

```
THIS SOFTWARE IS PROVIDED BY THE AUTHOR AND CONTRIBUTORS ``AS IS'' AND ANY
EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED.  IN NO EVENT SHALL THE AUTHOR OR CONTRIBUTORS BE LIABLE FOR
ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
```
