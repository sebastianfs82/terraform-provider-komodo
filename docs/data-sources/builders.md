---
page_title: "komodo_builders Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo builders visible to the authenticated user.
---

# komodo_builders (Data Source)

Lists all Komodo builders visible to the authenticated user.

## Example Usage

```terraform
data "komodo_builders" "all" {}

output "builder_names" {
  value = [for b in data.komodo_builders.all.builders : b.name]
}
```

## Schema

### Read-Only

- `builders` (List of Object) The list of builders. (see [below for nested schema](#nestedatt--builders))

<a id="nestedatt--builders"></a>
### Nested Schema for `builders`

Read-Only:

- `id` (String) The builder identifier (ObjectId).
- `name` (String) The name of the builder.
- `builder_type` (String) The builder type (`Server`, `Aws`, or `Url`).
