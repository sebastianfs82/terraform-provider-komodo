---
page_title: "komodo_server_prune_buildx Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Prunes the docker buildx cache on the target server.
---

# komodo_server_prune_buildx (Action)

Prunes the docker buildx cache on the target server.

## Example Usage

```shell
terraform apply -invoke action.komodo_server_prune_buildx.example
```

```terraform
resource "komodo_server" "my_server" {
  name = "my-server"
  # ...
}

action "komodo_server_prune_buildx" "example" {
  config {
    server = komodo_server.my_server.name
  }
}
```

## Schema

### Config

#### Required

- `server` (String) Id or name of the server on which to prune the buildx cache.
