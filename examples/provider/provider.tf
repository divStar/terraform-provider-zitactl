provider "zitactl" {
  domain                = "zitadel.example.com"
  service_account_key   = base64decode(module.zitadel.machine_user_key)
  skip_tls_verification = true
}