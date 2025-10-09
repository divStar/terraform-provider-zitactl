data "zitactl_orgs" "this" {
  name        = var.zitadel_orga_name
  name_method = "TEXT_QUERY_METHOD_STARTS_WITH_IGNORE_CASE"
}