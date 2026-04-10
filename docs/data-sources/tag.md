---
page_title: "komodo_tag Data Source - komodo"
subcategory: ""
description: |-
  Reads an existing Komodo tag.
---

# komodo_tag (Data Source)

Reads an existing Komodo tag.

## Example Usage

```terraform
data "komodo_tag" "example" {
  name = "my-tag"
}
```

## Schema

### Required

- `name` (String) The tag name.

### Read-Only

- `id` (String) The tag identifier (ObjectId).
- `color` (String) The tag color.
- `owner` (String) The user ID of the tag owner.
