---
page_title: "komodo_start_deployment Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Starts the target deployment.
---

# komodo_start_deployment (Action)

Starts the target deployment.

## Example Usage

```shell
terraform apply -invoke action.komodo_start_deployment.example
```

```terraform
resource "komodo_stack" "my_stack" {
  name = "my-stack"
  # ...
}

action "komodo_start_deployment" "example" {
  config {
    deployment = "my-deployment"
  }
}
```

## Schema

### Config

#### Required

- `deployment` (String) Id or name of the deployment to start.
