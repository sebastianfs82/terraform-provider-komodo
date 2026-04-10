resource "komodo_provider_account" "example" {
  domain        = "github.com"
  https_enabled = true
  username      = "myuser"
  token         = var.github_token
}
