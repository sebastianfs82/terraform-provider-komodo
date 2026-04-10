resource "komodo_network" "example" {
  server_id = komodo_server.example.id
  name      = "my-network"
}
