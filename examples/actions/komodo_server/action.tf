resource "komodo_server" "example" {
  name    = "prod-server"
  address = "https://my-server:8120"
}

# Prune unused docker images.
action "komodo_server" "prune_images" {
  config {
    id     = komodo_server.example.id
    action = "prune_images"
  }
}

# Prune stopped containers.
action "komodo_server" "prune_containers" {
  config {
    id     = komodo_server.example.id
    action = "prune_containers"
  }
}

# Run a full docker system prune (including volumes).
action "komodo_server" "prune_system" {
  config {
    id     = komodo_server.example.id
    action = "prune_system"
  }
}
