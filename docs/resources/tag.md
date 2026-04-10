---
page_title: "komodo_tag Resource - komodo"
subcategory: ""
description: |-
  Manages a Komodo tag.
---

# komodo_tag (Resource)

Manages a Komodo tag.

## Example Usage

```terraform
resource "komodo_tag" "example" {
  name  = "my-tag"
  color = "Red"
}
```

## Schema

### Required

- `name` (String) The tag name. Changing this value forces recreation of the resource.

### Optional

- `color` (String) The tag color. Valid values include: `Amber`, `Blue`, `Cyan`, `DarkAmber`, `DarkBlue`, `DarkCyan`, `DarkEmerald`, `DarkFuchsia`, `DarkGreen`, `DarkIndigo`, `DarkLime`, `DarkOrange`, `DarkPink`, `DarkPurple`, `DarkRed`, `DarkRose`, `DarkSky`, `DarkSlate`, `DarkTeal`, `DarkViolet`, `DarkYellow`, `Emerald`, `Fuchsia`, `Green`, `Indigo`, `LightAmber`, `LightBlue`, `LightCyan`, `LightEmerald`, `LightFuchsia`, `LightGreen`, `LightIndigo`, `LightLime`, `LightOrange`, `LightPink`, `LightPurple`, `LightRed`, `LightRose`, `LightSky`, `LightSlate`, `LightTeal`, `LightViolet`, `LightYellow`, `Lime`, `Orange`, `Pink`, `Purple`, `Red`, `Rose`, `Sky`, `Slate`, `Teal`, `Violet`, `Yellow`.

### Read-Only

- `id` (String) The tag identifier (ObjectId).
- `owner` (String) The user ID of the tag owner (set automatically by the API).

## Import

Import is supported using the following syntax:

```shell
terraform import komodo_tag.example my-tag
```
