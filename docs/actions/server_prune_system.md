---
page_title: "komodo_server_prune_system Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Prunes the docker system on the target server, including volumes.
---

# komodo_server_prune_system (Action)

Prunes the docker system on the target server, including volumes.

## Example Usage

```shell
terraform apply -invoke action.komodo_server_prune_system.example
```

```terraform
resource "komodo_server" "my_server" {
  name = "my-server"
  # ...
}

action "komodo_server_prune_system" "example" {
  config {
    server = komodo_server.my_server.name
  }
}
```

## Schema

### Config

#### Required

- `server` (String) Id or name of the server on which to prune the docker system.
