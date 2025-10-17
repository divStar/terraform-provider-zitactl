![GitHub License](https://img.shields.io/github/license/divStar/terraform-provider-zitactl?style=flat&color=pink)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/divStar/terraform-provider-zitactl/total?style=flat&color=peachpuff)
![Tests passed](https://github.com/divStar/terraform-provider-zitactl/actions/workflows/test.yml/badge.svg)

# `terraform-provider-zitactl` Terraform Provider for Zitadel v4

## Thanks

Many thanks to the [Zitadel](https://zitadel.com) team for their great work on Zitadel, the Zitadel Go library and
their Terraform provider! I am happy to be able to use these applications free of charge at home.

## Introduction

This is a Terraform provider for Zitadel v4. It uses the most recent `zitadel-go` Go library for Zitadel access via gRPC.
Its main feature is the **deferred client configuration**, which Zitadel currently sadly does not support yet
(see this [issue](https://github.com/zitadel/terraform-provider-zitadel/issues/167)).

The idea of this Terraform provider is to allow the user to install Zitadel in e.g. one module and reuse the machine key
(`zitadel-admin-sa.json`) directly in another. This allows it to install Zitadel and configure OIDC/OAuth2 access
in follow-up modules **in the same** Terraform run.

> [!IMPORTANT]
> If you do not intend to install Zitadel and configure OIDC/OAuth2 authentication in follow-up modules in one Terraform run,
> **do not use this provider**. Use Zitadel's official [Terraform provider](https://registry.terraform.io/providers/zitadel/zitadel/) instead.

The main reason for this recommendation is the unbelievable amount of data sources and resources the official Zitadel provider
offers - especially in contrast to this one.

Apart from the deferred client configuration (**not** provider configuration) and the fact, that the `terraform-plugin-framework`
is used (it is the recommended way to write new Terraform providers), this provider offers only the following
data sources and resources:
- ![data-source](https://img.shields.io/badge/data_source-blue?style=flat) Organization list ([`zitactl_org`](./docs/data-sources/orgs.md)),
- ![resource](https://img.shields.io/badge/resource-purple?style=flat) Project ([`zitactl_project`](./docs/resources/project.md)),
- ![resource](https://img.shields.io/badge/resource-purple?style=flat) Application OIDC ([`zitactl_application_oidc`](./docs/resources/application_oidc.md))

This makes it possible to install Zitadel in one module and configure one or several applications (e.g. pgAdmin v4)
with OIDC/OAuth2 authentication in one go.

> [!CAUTION]
> This provider is currently at an early stage. I have used it **on my MacBook** and the acceptance tests work in CI, but I might *not* have forseen all edge cases.
> **Use at your own risk!**

> [!WARNING]
> If you e.g. install Zitadel in one module and configure OIDC/OAuth2 authentication in another module, you might run into
> a situation where Zitadel **is not yet protected by a valid SSL certificate** (e.g. Traefik v3 usually issues its default
> self-signed certificate until it is able to acquire a valid certificate from a CA using ACME;
> this intermediate certificate is usually invalid).
> 
> To avoid errors like:
> ```
> Error: Error creating project: rpc error: code = Unavailable desc = connection error: desc = "transport: authentication handshake failed: x509: certificate signed by unknown authority"
> ```
> you either need to implement a way to wait until a proper certificate has been issued or set `skip_tls_verification` to `true` in the provider configuration.
> This might be a security risk though, but works for my homelab setting just fine.

## Requirements

- at least ![Terraform v1.5.7](https://img.shields.io/badge/Terraform-1.5.7-orange?logo=terraform) or ![OpenTofu 1.10.5](https://img.shields.io/badge/Terraform-1.10.5-peachpuff?logo=opentofu)
- ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/divStar/terraform-provider-zitactl?style=flat&logo=go)

## Using the provider

Here is an example, that I use in my Terraform script.

### Module `main.tf`

```hcl
local {
  # ...
  # We expect there to only be one company; this *should* always be the case on a fresh Zitadel installation.
  sanctum_orga_id = data.zitactl_orgs.this.ids[0]
}

# This data source retrieves the organization ID of `zitadel_orga_name`.
data "zitactl_orgs" "this" {
  name = var.zitadel_orga_name
}

# Creates the `pgadmin` project within the given `var.zitadel_orga_name` organization.
resource "zitactl_project" "this" {
  name                   = "pgadmin"
  org_id                 = local.sanctum_orga_id
  project_role_assertion = true
  project_role_check     = true
}

resource "zitactl_application_oidc" "this" {
  project_id = zitactl_project.this.id

  name                      = "pgadmin"
  redirect_uris             = ["https://pgadmin.${var.cluster.domain}/oauth2/authorize"]
  response_types            = ["OIDC_RESPONSE_TYPE_CODE"]
  grant_types               = ["OIDC_GRANT_TYPE_AUTHORIZATION_CODE"]
  app_type                  = "OIDC_APP_TYPE_WEB"
  auth_method_type          = "OIDC_AUTH_METHOD_TYPE_BASIC"
  post_logout_redirect_uris = ["https://pgadmin.${var.cluster.domain}/"]
}

# Installs [pgAdmin 4](https://github.com/rowanruseler/helm-charts/tree/main/charts/pgadmin4),
# a web-based administration tool for PostgreSQL.
module "pgadmin" {
  # ...
  pre_install_resources = [
    {
      yaml = templatefile("${path.module}/files/pgadmin.secret.env.pre-install.yaml.tftpl", {
        pgadmin_namespace    = local.pgAdmin.namespace
        pgadmin_secret_name  = var.pgadmin_secret_name
        oauth2_client_id     = zitactl_application_oidc.this.client_id
        oauth2_client_secret = zitactl_application_oidc.this.client_secret
      })
    },
    # ...
  ]
}
```

### Provider configuration `provider.tf`

```hcl
provider "zitactl" {
  domain = "zitadel.${var.cluster.domain}"
  service_account_key = base64decode(module.zitadel.machine_user_key)
  skip_tls_verification = true
}
```

The beauty of the provider configuration is, that variables and outputs **do not have to be known in advance**.
The client configuration is deferred until the first data source or resource is created. Only if the configuration
fails at that time, because e.g. some variables could not be resolved, an error is thrown.

## Developing the Provider

Make sure you have Go installed on your system and the version matches the version specified in the `go.mod` file.
It is also advisable to have `make` installed in order to use targets defined in the `GNUmakefile`.

For running acceptance tests, you'll need Docker (or Docker Desktop) installed and running, as the tests require a local Zitadel instance.

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

##### Testing against a real running Zitadel instance with valid SSL certificates

```bash
ZITACTL_DOMAIN=zitadel.my.domain \
ZITACTL_SKIP_TLS_VERIFICATION=false \
ZITACTL_SERVICE_ACCOUNT_KEY='{"type":"serviceaccount",...}' \
make testacc
```

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

## Tested configuration

This Terraform provider has been tested with the following configuration in CI/CD and local development:
- ![Terraform v1.5.7](https://img.shields.io/badge/Terraform-1.5.7-orange?logo=terraform) / ![OpenTofu 1.10.5](https://img.shields.io/badge/Terraform-1.10.5-peachpuff?logo=opentofu)
- <img src="https://raw.githubusercontent.com/homarr-labs/dashboard-icons/refs/heads/main/svg/zitadel-light.svg" width="16"> Zitadel **v4.3.3**
- ![PostgreSQL 17](https://img.shields.io/badge/PostgreSQL-17-lightcyan?logo=postgresql)

The versions above match those defined in `./tools/docker-compose.yml` used for acceptance testing.

This provider is also used in production in my [divStar/homelab project](https://github.com/divStar/homelab) for managing Zitadel authentication.

## Known issues:

* Acceptance tests do not completely cover all error cases, just the most likely ones.
* Only one data source and two sources are supported. (will not change for now unless someone creates PRs)
