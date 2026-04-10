---
page_title: "komodo_stack_deploy Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Triggers a docker compose up on the target Komodo stack.
---

# komodo_stack_deploy (Action)

Triggers a `docker compose up` on the target Komodo stack.

## Example Usage

Actions are not executed automatically — they must be triggered either by an
`action_trigger` block in a resource lifecycle, or invoked explicitly on the CLI:

```shell
terraform apply -invoke action.komodo_stack_deploy.example
```

### Trigger after stack resource changes

```terraform
resource "komodo_stack" "my_app" {
  name = "my-app"
  # ...
}

action "komodo_stack_deploy" "example" {
  config {
    stack = komodo_stack.my_app.name
  }
}

# Automatically deploy whenever the stack resource is created or updated.
resource "terraform_data" "deploy_trigger" {
  input = komodo_stack.my_app.id

  lifecycle {
    action_trigger {
      events  = [after_create, after_update]
      actions = [action.komodo_stack_deploy.example]
    }
  }
}
```

### Deploy only specific services

```terraform
action "komodo_stack_deploy" "partial" {
  config {
    stack    = "my-app"
    services = ["web", "api"]
  }
}
```

## Schema

### Config

#### Required

- `stack` (String) Id or name of the stack to deploy.

#### Optional

- `services` (List of String) Filter to only deploy specific services. If empty, deploys all services.
- `stop_time` (Number) Override the default termination max time in seconds. Only used if the stack needs to be taken down first.
