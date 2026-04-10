---
page_title: "komodo_server_prune_containers Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Prunes the docker containers on the target server.
---

# komodo_server_prune_containers (Action)

Prunes the docker containers on the target server.

## Example Usage

```shell
terraform apply -invoke action.komodo_server_prune_containers.example
```

```terraform
resource "komodo_server" "my_server" {
  name = "my-server"
  # ...
}

action "komodo_server_prune_containers" "example" {
  config {
    server = komodo_server.my_server.name
  }
}
```

## Schema

### Config

#### Required

- `server` (String) Id or name of the server on which to prune containers.
