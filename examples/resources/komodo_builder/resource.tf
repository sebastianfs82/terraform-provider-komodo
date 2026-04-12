resource "komodo_builder" "server" {
  name = "my-builder"
  type = "Server"

  server_config = {
    server_id = komodo_server.example.id
  }
}

resource "komodo_builder" "url" {
  name = "peripheral-builder"
  type = "Url"

  url_config = {
    address = "https://periphery.example.com"
  }
}
