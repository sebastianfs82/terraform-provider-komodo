resource "komodo_builder" "server" {
  name         = "my-builder"
  builder_type = "Server"

  server_config = {
    server_id = komodo_server.example.id
  }
}

resource "komodo_builder" "url" {
  name         = "peripheral-builder"
  builder_type = "Url"

  url_config = {
    address = "https://periphery.example.com"
  }
}
