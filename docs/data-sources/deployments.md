---
page_title: "komodo_deployments Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo deployments visible to the authenticated user.
---

# komodo_deployments (Data Source)

Lists all Komodo deployments visible to the authenticated user.

## Example Usage

```terraform
data "komodo_deployments" "all" {}

output "deployment_names" {
  value = [for d in data.komodo_deployments.all.deployments : d.name]
}
```

## Schema

### Read-Only

- `deployments` (List of Object) The list of deployments. (see [below for nested schema](#nestedatt--deployments))

<a id="nestedatt--deployments"></a>
### Nested Schema for `deployments`

Read-Only:

- `id` (String) The deployment identifier (ObjectId).
- `name` (String) The name of the deployment.
- `server_id` (String) The ID of the server the deployment runs on.
- `image` (String) The container image (or build ID for build-backed deployments).
- `send_alerts` (Boolean) Whether alerts are enabled for this deployment.
