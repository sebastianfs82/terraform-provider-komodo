resource "komodo_server" "example" {
  name    = "my-server"
  address = "https://192.168.1.100:8120"
}

resource "komodo_server" "production" {
  name                  = "prod-server"
  address               = "https://prod.example.com:8120"
  region                = "us-east"
  enabled               = true
  send_cpu_alerts       = true
  send_mem_alerts       = true
  send_disk_alerts      = true
  cpu_warning           = 80.0
  cpu_critical          = 95.0
  mem_warning           = 80.0
  mem_critical          = 95.0
}
