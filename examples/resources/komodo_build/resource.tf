resource "komodo_build" "example" {
  name       = "my-service"
  builder_id = komodo_builder.example.id

  source {
    path   = "myorg/my-service"
    branch = "main"
  }

  image {
    name = "myorg/my-service"

    registry {
      account_id = komodo_registry_account.example.id
    }
  }

  build {
    path = "."

    argument {
      name  = "BUILD_ENV"
      value = "production"
    }
    argument {
      name           = "API_KEY"
      value          = "my-secret-value"
      secret_enabled = true
    }
  }
}
