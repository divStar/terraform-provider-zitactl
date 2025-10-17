## Developing the Provider

### Introduction

Make sure you have **Go** installed on your system, and the version matches the version specified in the `go.mod` file.
It is also advisable to have **`make`** installed to use targets defined in the `GNUmakefile`.

To run the acceptance tests, you will need **Docker** (or Docker Desktop) up and running, as the tests require a local Zitadel instance.

### `GNUmakefile` targets

#### `default` target

Runs the standard development workflow: formats code, runs linters, and builds the provider.
```bash
make
# or explicitly:
make default
```

This is equivalent to running `make fmt lint build`.

#### `build` target

Compiles and builds the provider to verify that all packages compile successfully.

> [!NOTE]
> This target does not produce a binary. Use `make artifact` to create a binary in the root directory of the project.

#### `artifact` target

Compiles and builds the provider with debug flags (`-gcflags="all=-N -l"`) and produces a binary in the root directory of the project.

This target is useful for building a binary for local testing or debugging.

#### `lint` target

Runs `golangci-lint` to check code quality and style issues.

#### `generate` target

Generates copyright headers (using `copywrite`) and Terraform provider documentation (using `tfplugindocs`).

#### `fmt` target

Formats all Go code in the project using `gofmt`.

#### `test` target

Runs unit tests with coverage reporting.
```bash
make test
```

#### `testacc` target

Runs acceptance tests against a Zitadel instance.

> [!NOTE]
> Requires a running Zitadel instance (see [Test Infrastructure](#test-infrastructure) below).

```bash
make testacc
```

**Environment variables:**
- `ZITACTL_DOMAIN` - the domain of your Zitadel instance (default: `localhost`)
- `ZITACTL_SKIP_TLS_VERIFICATION` - set to `true` to skip TLS certificate verification (default: `false`). Required when using the local Docker setup.
- `ZITACTL_SERVICE_ACCOUNT_KEY` - the machine key JSON for authenticating with Zitadel. If not set, the target will attempt to read from `./tools/serviceaccount/zitadel-admin-sa.json`

**Examples:**

##### Using the local Docker setup (default)

```bash
make zitadel-up
make testacc
make zitadel-down
```

> [!NOTE]
> `ZITACTL_SERVICE_ACCOUNT_KEY` will be acquired automatically from the local Docker setup.
> `ZITACTL_SKIP_TLS_VERIFICATION` is set to `true` by default.
> `ZITACTL_DOMAIN` is set to `localhost` by default.`

##### Testing against a real running Zitadel instance with valid SSL certificates

```bash
ZITACTL_DOMAIN=zitadel.my.domain \
ZITACTL_SKIP_TLS_VERIFICATION=false \
ZITACTL_SERVICE_ACCOUNT_KEY='{"type":"serviceaccount","keyId":"342080798231035907","key":"-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA1WfbTxO8T7Vkg7XC4zFnJ3xz62/Lgfnkli3/J6WhZbysIIoz\n8tazu...\nn60G5I1YNdwwU2BckLx3gL2v3vtx9p0HQD/l29Mts7Q2ZcJxNxm9ZQ==\n-----END RSA PRIVATE KEY-----\n","expirationDate":"9999-12-31T23:59:59Z","userId":"342080798230904835"}' \
make testacc
```

> [!NOTE]
> Be careful how you assign a value to `ZITACTL_SERVICE_ACCOUNT_KEY`, because your shell might interpolate special
> characters (e.g. `\n`). Check using `printf '%s\n' "$ZITACTL_SERVICE_ACCOUNT_KEY"` to see if the value is correct.

> [!WARNING]
> `PAT`s or any other authentication methods are **not supported**.
> Feel free to create a ticket or (preferably) a PR if you need this feature.

#### `test-release` target

Simulates a GoReleaser release locally using `--snapshot` mode.
Useful for debugging release configuration issues before pushing a tag.

> [!NOTE]
> Requires the `GPG_FINGERPRINT` environment variable to be set.
> Have a look at the `./get-gpg-passphrase.sh` script to see how to set up GPG with this project.

#### Test Infrastructure

The following targets manage a local Docker Compose stack that provides PostgreSQL and Zitadel
for development and acceptance testing.

##### `zitadel-up` target

Starts the local Zitadel test infrastructure (PostgreSQL and Zitadel) in Docker.
```bash
make zitadel-up
```

This is required before running acceptance tests with `make testacc` and it's what the `test.yml` CI pipeline uses.

##### `zitadel-down` target

Stops and tears down the local Zitadel test infrastructure.
```bash
make zitadel-down
```

##### `zitadel-logs` target

Follows the logs of the running Zitadel Docker Compose stack.
```bash
make zitadel-logs
```

This can be useful if you have to debug some odd behavior.

### Typical Development Workflow

1. Create a new feature branch.
2. Make your changes to the code.
3. Run `make` to format, lint, build, and generate documentation
4. Unit tests do not require a running Zitadel instance, therefore the following command suffices:
    ```bash
       make test
    ```
5. Develop and run acceptance tests (they are **not** optional):
    ```bash
    make zitadel-up      # Start test infrastructure
    make testacc         # Run acceptance tests
    make zitadel-down    # Clean up when done
    ```

### Adding or updating dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

### Creating a PR

1. Fork the project.
2. Implement your fix or feature.
3. Open a PR against the `master` branch of this repository.
4. Test your changes locally using the various `GNUmakefile` targets.
5. Add a link to the issue in the PR description (if applicable).
6. Describe your changes in the PR description.
7. Create or update unit or acceptance tests if you changed the source code.
8. Update the provider documentation and examples if applicable.
9. Wait for the PR to be reviewed and merged.
10. Once merged, at some point a new release will be created.

### Releasing a new version

To release a new version of the provider, it's usually enough to create a new tag in the repository.
The CI pipeline will automatically build and publish a new release.

If the CI pipeline fails and the error is not obvious, one can try to debug it locally using the `test-release` target (`GNUmakefile`).

GPG key (`GPG_PRIVATE_KEY`) and passphrase (`GPG_PASSPHRASE`) have been set up for this repository.

Under the hood, the `test-release` target uses the `goreleaser` tool to build and publish a new release.
Have a look at the [documentation](https://goreleaser.com/customization/release/) for more information.