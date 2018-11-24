# Contributing

When contributing to this repository, please first discuss the change you wish to make via issue, email, or any other method with the owners of this repository before making a change. This usually saves time & effort.

Please note we have a [code of conduct](./CODE_OF_CONDUCT.md), please follow it in all your interactions with the project.

## Development

Gitzup uses `Make` as a build system, though internally the standard Go tools are used (`go build`, `go install`, `go test`, etc).

### Building

```bash
$ make
```

### Updating GCP service account key for Travis CI

To update the GCP service account JSON key file, do the following:

1. Obtain an update GCP service account JSON key from the GCP console and store it as `.travis-ci-sa-key.json` in the repository root directory.

2. Run the following:

    ```
    $ travis encrypt-file .travis-ci-sa-key.json
    encrypting .travis-ci-sa-key.json for kfirz/gitzup
    storing result as .travis-ci-sa-key.json.enc
    storing secure env variables for decryption

    Please add the following to your build script (before_install stage in your .travis.yml, for instance):

    openssl aes-... -K $encrypted_..._key -iv $encrypted_..._iv -in .travis-ci-sa-key.json.enc -out .travis-ci-sa-key.json -d

    Pro Tip: You can add it automatically by running with --add.

    Make sure to add .travis-ci-sa-key.json.enc to the git repository.
    Make sure not to add .travis-ci-sa-key.json to the git repository.
    Commit all changes to your .travis.yml.
    ```
   
   **NOTE:** keep the output open! you will need to copy some of it in step 3!
  
3. Open the `.tracvis.yml` file and update the `openssl` command with the updated form printed in step 2.

### Testing

```bash
$ make test
```

## Pull Request process

Aside from the actual change in source code, please ensure your PR update any relevant tests and/or adds new tests as necessary. PRs that lower test coverage, or cause test failures, will not be accepted.

For cases where the change affects information displayed in the documentation, please ensure the PR updates the documentation as well (eg. `README.md`).

## Releasing

TBD.
