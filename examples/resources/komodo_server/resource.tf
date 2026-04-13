# Minimal server
resource "komodo_server" "example" {
  name    = "my-server"
  address = "wss://192.168.1.100:8120"
}

# Server with alerts, maintenance window, and extra options
resource "komodo_server" "production" {
  name                                 = "prod-server"
  address                              = "wss://prod.example.com:8120"
  region                               = "us-east"
  enabled                              = true
  certificate_verification_enabled     = true
  auto_prune_images_enabled            = true
  auto_rotate_keys_enabled             = true
  historical_system_statistics_enabled = true

  alerts {
    enabled = true
    types   = ["cpu", "memory", "disk", "unreachable"]
    thresholds {
      cpu_warning     = 75.0
      cpu_critical    = 90.0
      memory_warning  = 75.0
      memory_critical = 90.0
      disk_warning    = 80.0
      disk_critical   = 95.0
    }
  }

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
