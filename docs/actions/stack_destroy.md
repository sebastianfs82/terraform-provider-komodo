---
page_title: "komodo_stack_destroy Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Triggers a docker compose down on the target Komodo stack.
---

# komodo_stack_destroy (Action)

Triggers a `docker compose down` on the target Komodo stack.

## Example Usage

Actions are not executed automatically — they must be triggered either by an
`action_trigger` block in a resource lifecycle, or invoked explicitly on the CLI:

```shell
terraform apply -invoke action.komodo_stack_destroy.example
```

### Trigger to bring down a stack on demand

```terraform
resource "komodo_stack" "my_app" {
  name = "my-app"
  # ...
}

action "komodo_stack_destroy" "example" {
  config {
    stack = komodo_stack.my_app.name
  }
}

# Change `input` to a new value (e.g. a timestamp) and re-apply to bring down again.
resource "terraform_data" "destroy_trigger" {
  input = komodo_stack.my_app.id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.komodo_stack_destroy.example]
    }
  }
}
```

### Bring down specific services with options

```terraform
action "komodo_stack_destroy" "full" {
  config {
    stack          = "my-app"
    services       = ["web", "api"]
    remove_orphans = true
    stop_time      = 30
  }
}
```

## Schema

### Config

#### Required

- `stack` (String) Id or name of the stack to destroy.

#### Optional

- `services` (List of String) Filter to only destroy specific services. If empty, destroys all services.
- `remove_orphans` (Boolean) Pass `--remove-orphans` to `docker compose down`.
- `stop_time` (Number) Override the default termination max time in seconds.
