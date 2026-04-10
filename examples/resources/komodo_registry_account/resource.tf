resource "komodo_registry_account" "example" {
  domain   = "docker.io"
  username = "myuser"
  token    = var.dockerhub_token
}
