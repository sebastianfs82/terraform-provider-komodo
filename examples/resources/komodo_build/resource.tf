resource "komodo_build" "example" {
  name       = "my-service"
  builder_id = komodo_builder.example.id
  image_name = "myorg/my-service"
  repo       = "myorg/my-service"
  branch     = "main"
}
