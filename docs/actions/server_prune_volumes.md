---
page_title: "komodo_server_prune_volumes Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Prunes the docker volumes on the target server.
---

# komodo_server_prune_volumes (Action)

Prunes the docker volumes on the target server.

## Example Usage

```shell
terraform apply -invoke action.komodo_server_prune_volumes.example
```

```terraform
resource "komodo_server" "my_server" {
  name = "my-server"
  # ...
}

action "komodo_server_prune_volumes" "example" {
  config {
    server = komodo_server.my_server.name
  }
}
```

## Schema

### Config

#### Required

- `server` (String) Id or name of the server on which to prune volumes.
