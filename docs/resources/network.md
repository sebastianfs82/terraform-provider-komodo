---
page_title: "komodo_network Resource - komodo"
subcategory: ""
description: |-
  Creates a docker network on a Komodo-managed server.
---

# komodo_network (Resource)

Creates a docker network on a Komodo-managed server by running `docker network create {name}`.

Changing `server_id` or `name` forces a new resource.

> **Note:** The Komodo API has no endpoint to delete docker networks. Destroying this resource removes it from Terraform state, but the network on the server is **not** removed. You must delete the network manually if required.

## Example Usage

```terraform
resource "komodo_network" "example" {
  server_id = "my-server"
  name      = "my-network"
}
```

## Schema

### Required

- `name` (String) The name of the docker network to create. Changing this value forces recreation of the resource.
- `server_id` (String) The server ID or name on which to create the network. Changing this value forces recreation of the resource.

### Read-Only

- `id` (String) The resource identifier in the form `server_id:name`.

## Import

Import an existing docker network by providing its composite ID in the form `server_id:name`:

```shell
terraform import komodo_network.example my-server:my-network
```
