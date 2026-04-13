resource "komodo_repo" "example" {
  name      = "my-repo"
  server_id = komodo_server.example.id

  source {
    path   = "myorg/myrepo"
    branch = "main"
  }
}
