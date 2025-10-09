# Examples

This directory contains examples that are mostly used for documentation, but can also be run/tested manually via the Terraform CLI.

The document generation tool looks for files in the following locations by default. All other *.tf files besides the ones mentioned below are ignored by the documentation tool. This is useful for creating examples that can run and/or are testable even if some parts are not relevant for the documentation.

* **provider/provider.tf** example file for the provider index page
* **data-sources/`zitactl_org`/data-source.tf** example file for the named data source page
* **resources/`zitactl_project`/resource.tf** example file for the named data source page
* **resources/`zitactl_application_oidc`/resource.tf** example file for the named data source page
