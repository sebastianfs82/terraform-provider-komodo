resource "komodo_stack" "example" {
  name      = "my-stack"
  server_id = komodo_server.example.id

  compose = {
    contents = file("${path.module}/compose.yaml")
  }
}
