resource "komodo_user" "example" {
  username              = "alice"
  password              = "securePassword1!"
  create_server_enabled = true
  create_build_enabled  = true
}
