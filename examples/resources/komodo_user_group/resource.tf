resource "komodo_user_group" "example" {
  name  = "developers"
  users = [komodo_user.alice.id, komodo_user.bob.id]
}

resource "komodo_user_group" "everyone" {
  name             = "Everyone"
  everyone_enabled = true
}
