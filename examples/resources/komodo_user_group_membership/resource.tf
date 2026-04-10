resource "komodo_user_group_membership" "example" {
  user_group = komodo_user_group.example.id
  user       = komodo_user.alice.id
}
