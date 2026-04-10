resource "komodo_deployment" "example" {
  name      = "my-service"
  server_id = komodo_server.example.id

  image = {
    type  = "Image"
    image = "nginx:latest"
  }
}

resource "komodo_deployment" "from_build" {
  name      = "my-built-service"
  server_id = komodo_server.example.id

  image = {
    type     = "Build"
    build_id = komodo_build.example.id
  }
}
