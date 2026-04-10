---
page_title: "komodo_resource_syncs Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo resource syncs visible to the authenticated user.
---

# komodo_resource_syncs (Data Source)

Lists all Komodo resource syncs visible to the authenticated user.

## Example Usage

```terraform
data "komodo_resource_syncs" "all" {}

output "sync_names" {
  value = [for s in data.komodo_resource_syncs.all.resource_syncs : s.name]
}
```

## Schema

### Read-Only

- `resource_syncs` (List of Object) The list of resource syncs. (see [below for nested schema](#nestedatt--resource_syncs))

<a id="nestedatt--resource_syncs"></a>
### Nested Schema for `resource_syncs`

Read-Only:

- `id` (String) The resource sync identifier (ObjectId).
- `name` (String) The name of the resource sync.
- `repo` (String) The git repository path (owner/repo).
- `branch` (String) The git branch.
- `webhook_enabled` (Boolean) Whether webhook triggers are enabled.
- `managed` (Boolean) Whether the sync manages resources (creates/deletes).
