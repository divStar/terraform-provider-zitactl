![GitHub License](https://img.shields.io/github/license/divStar/terraform-provider-zitactl?style=flat&color=pink)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/divStar/terraform-provider-zitactl/total?style=flat&color=peachpuff)
![Tests passed](https://github.com/divStar/terraform-provider-zitactl/actions/workflows/test.yml/badge.svg)

# `terraform-provider-zitactl` Terraform Provider for Zitadel v4

## Thank you!

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

## Tested configuration

This Terraform provider has been tested with the following configuration in CI/CD and local development:
- ![Terraform v1.5.7](https://img.shields.io/badge/Terraform-1.5.7-orange?logo=terraform) / ![OpenTofu 1.10.5](https://img.shields.io/badge/Terraform-1.10.5-peachpuff?logo=opentofu)
- <img src="https://raw.githubusercontent.com/homarr-labs/dashboard-icons/refs/heads/main/svg/zitadel-light.svg" width="16"> Zitadel **v4.3.3**
- ![PostgreSQL 17](https://img.shields.io/badge/PostgreSQL-17-lightcyan?logo=postgresql)

The versions above match those defined in `./tools/docker-compose.yml` used for acceptance testing.

This provider is also used in production in my [divStar/homelab project](https://github.com/divStar/homelab) for managing Zitadel authentication.

## Known issues:

* Limited amount of data sources and resources.
* #7 fixes the issue of `Read` functions being called during a resource state refresh or early data-source read.
This should make it possible to e.g. run `tofu apply` (or `tofu plan`) a second time to update everything without issues.