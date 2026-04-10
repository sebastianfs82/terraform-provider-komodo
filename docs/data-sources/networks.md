---
page_title: "komodo_networks Data Source - komodo"
subcategory: ""
description: |-
  Lists all docker networks on a Komodo-managed server.
---

# komodo_networks (Data Source)

Lists all docker networks on a Komodo-managed server.

## Example Usage

```terraform
data "komodo_networks" "example" {
  server_id = "my-server"
}

output "network_names" {
  value = [for n in data.komodo_networks.example.networks : n.name]
}
```

## Schema

### Required

- `server_id` (String) The server ID or name to list networks on.

### Read-Only

- `networks` (List of Object) The list of docker networks on the server. (see [below for nested schema](#nestedatt--networks))

<a id="nestedatt--networks"></a>
### Nested Schema for `networks`

Read-Only:

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
- `name` (String) The name of the docker network.
- `network_id` (String) The docker-assigned network ID.
- `scope` (String) The scope of the network (e.g. `local`, `swarm`).
- `server_id` (String) The server ID or name the network belongs to.
