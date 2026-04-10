---
page_title: "komodo_actions Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo actions visible to the authenticated user.
---

# komodo_actions (Data Source)

Lists all Komodo actions visible to the authenticated user.

## Example Usage

```terraform
data "komodo_actions" "all" {}

output "action_names" {
  value = [for a in data.komodo_actions.all.actions : a.name]
}
```

## Schema

### Read-Only

- `actions` (List of Object) The list of actions. (see [below for nested schema](#nestedatt--actions))

<a id="nestedatt--actions"></a>
### Nested Schema for `actions`

Read-Only:

- `id` (String) The action identifier (ObjectId).
- `name` (String) The name of the action.
- `file_contents` (String) The Deno TypeScript file contents of the action.
- `run_at_startup` (Boolean) Whether the action runs at Komodo startup.
- `schedule` (String) The cron schedule for the action.
- `schedule_enabled` (Boolean) Whether the schedule is enabled.
- `webhook_enabled` (Boolean) Whether webhook triggers are enabled.
