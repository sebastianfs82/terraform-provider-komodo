---
page_title: "komodo_alerters Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo alerters visible to the authenticated user.
---

# komodo_alerters (Data Source)

Lists all Komodo alerters visible to the authenticated user.

## Example Usage

```terraform
data "komodo_alerters" "all" {}

output "enabled_alerters" {
  value = [for a in data.komodo_alerters.all.alerters : a.name if a.enabled]
}
```

## Schema

### Read-Only

- `alerters` (List of Object) The list of alerters. (see [below for nested schema](#nestedatt--alerters))

<a id="nestedatt--alerters"></a>
### Nested Schema for `alerters`

Read-Only:

- `id` (String) The alerter identifier (ObjectId).
- `name` (String) The name of the alerter.
- `enabled` (Boolean) Whether the alerter is enabled.
- `endpoint_type` (String) The alerter endpoint type (`Slack`, `Discord`, `Custom`, `Ntfy`, or `Pushover`).
