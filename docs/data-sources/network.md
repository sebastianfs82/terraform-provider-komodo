---
page_title: "komodo_network Data Source - komodo"
subcategory: ""
description: |-
  Reads an existing docker network on a Komodo-managed server.
---

# komodo_network (Data Source)

Reads an existing docker network on a Komodo-managed server.

## Example Usage

```terraform
data "komodo_network" "example" {
  server_id = "my-server"
  name      = "my-network"
}

output "network_driver" {
  value = data.komodo_network.example.driver
}
```

## Schema

### Required

- `name` (String) The name of the docker network to look up.
- `server_id` (String) The server ID or name to query networks on.

### Read-Only

- `attachable` (Boolean) Whether manual container attachment is allowed.
- `created` (String) The timestamp when the network was created.
- `driver` (String) The network driver (e.g. `bridge`, `overlay`).
- `enable_ipv6` (Boolean) Whether IPv6 is enabled on the network.
- `in_use` (Boolean) Whether the network is currently attached to one or more containers.
- `ingress` (Boolean) Whether the network is an ingress network (used for swarm routing mesh).
- `internal` (Boolean) Whether the network is internal (not connected to the external network).
- `ipam_driver` (String) The IPAM driver used by the network.
- `ipam_gateway` (String) The IPAM gateway configured for the network.
- `ipam_subnet` (String) The IPAM subnet configured for the network.
- `network_id` (String) The docker-assigned network ID.
- `scope` (String) The scope of the network (e.g. `local`, `swarm`).
