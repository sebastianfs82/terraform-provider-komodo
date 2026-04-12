resource "komodo_user_group_membership" "example" {
  user_group_id = komodo_user_group.example.id
  user_id       = komodo_user.alice.id
}
