resource "komodo_stack" "example" {
  name      = "my-stack"
  server_id = komodo_server.example.id

  files = {
    contents = [
      {
        path     = "compose.yaml"
        contents = file("${path.module}/compose.yaml")
      }
    ]
  }
}
