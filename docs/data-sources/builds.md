---
page_title: "komodo_builds Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Lists all Komodo builds visible to the authenticated user.
---

# komodo_builds (Data Source)

Lists all Komodo builds visible to the authenticated user.

## Example Usage

```terraform
data "komodo_builds" "all" {}

output "build_names" {
  value = [for b in data.komodo_builds.all.builds : b.name]
}
```

## Schema

### Read-Only

- `builds` (List of Object) The list of builds. (see [below for nested schema](#nestedatt--builds))

<a id="nestedatt--builds"></a>
### Nested Schema for `builds`

Read-Only:

- `id` (String) The build identifier (ObjectId).
- `name` (String) The name of the build.
- `builder_id` (String) The ID of the builder used by this build.
- `image_name` (String) The target image name.
- `image_tag` (String) The target image tag.
- `repo` (String) The git repository path (owner/repo).
- `branch` (String) The git branch.
- `webhook_enabled` (Boolean) Whether webhook triggers are enabled.
