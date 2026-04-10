---
page_title: "komodo_onboarding_key Data Source - terraform-provider-komodo"
subcategory: ""
description: |-
  Reads an existing Komodo onboarding key by name or public key.
---

# komodo_onboarding_key (Data Source)

Reads an existing Komodo onboarding key by `name` or `public_key`. At least one must be set.
When `public_key` is provided alongside `name`, `public_key` takes precedence.

## Example Usage

```terraform
# Look up by name
data "komodo_onboarding_key" "by_name" {
  name = "my-onboarding-key"
}

# Look up by public key (e.g. after creating with the resource)
resource "komodo_onboarding_key" "prod" {
  name = "prod-key"
}

data "komodo_onboarding_key" "prod" {
  public_key = komodo_onboarding_key.prod.public_key
}
```

## Schema

### Optional

- `public_key` (String) The public key used to identify the onboarding key. If set alongside `name`, takes precedence.
- `name` (String) The name of the onboarding key.

### Read-Only

- `enabled` (Boolean) Whether the onboarding key is enabled.
- `expires` (Number) The expiry timestamp (Unix ms). `0` means no expiry.
- `privileged` (Boolean) Whether the onboarding key grants privileged access.
- `copy_server` (String) ID or name of a server to copy configuration from.
- `create_builder` (Boolean) Whether to create a builder for the onboarded server.
