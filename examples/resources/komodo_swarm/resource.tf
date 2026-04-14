# Minimal swarm
resource "komodo_swarm" "example" {
  name = "my-swarm"
}

# Swarm with server nodes, links, and a maintenance window
resource "komodo_swarm" "production" {
  name           = "prod-swarm"
  server_ids     = ["6627c3e4f1a2b3c4d5e6f7a8", "6627c3e4f1a2b3c4d5e6f7a9"]
  links          = ["http://portainer.example.com", "http://grafana.example.com"]
  alerts_enabled = true

  maintenance {
    name             = "weekly-patching"
    description      = "Saturday night patch window"
    schedule_type    = "Weekly"
    day_of_week      = "Saturday"
    hour             = 2
    minute           = 0
    duration_minutes = 120
    timezone         = "America/New_York"
    enabled          = true
  }
}

# Swarm referencing servers by resource
resource "komodo_server" "manager1" {
  name    = "swarm-manager-1"
  address = "wss://manager1.example.com:8120"
}

resource "komodo_server" "manager2" {
  name    = "swarm-manager-2"
  address = "wss://manager2.example.com:8120"
}

resource "komodo_swarm" "cluster" {
  name       = "prod-cluster"
  server_ids = [komodo_server.manager1.id, komodo_server.manager2.id]
}
