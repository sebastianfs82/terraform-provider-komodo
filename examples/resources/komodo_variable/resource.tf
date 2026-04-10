resource "komodo_variable" "example" {
  name  = "MY_VARIABLE"
  value = "my-value"
}

resource "komodo_variable" "secret" {
  name      = "MY_SECRET"
  value     = var.my_secret
  is_secret = true
}
