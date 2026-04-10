resource "komodo_user" "example" {
  username       = "alice"
  password       = "securePassword1!"
  create_servers = true
  create_builds  = true
}
