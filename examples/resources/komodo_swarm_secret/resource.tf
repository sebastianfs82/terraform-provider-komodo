resource "komodo_swarm_secret" "example" {
  swarm = "my-swarm"
  name  = "db-password"
  data  = "replace-with-secure-value"
}

# Optional secret driver and labels
resource "komodo_swarm_secret" "with_options" {
  swarm           = "my-swarm"
  name            = "api-token"
  data            = "replace-with-secure-token"
  driver          = ""
  template_driver = ""
  labels          = ["environment=production", "managed-by=terraform"]
}
