---
page_title: "komodo_pull_deployment Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Pulls the latest image for the target deployment.
---

# komodo_pull_deployment (Action)

Pulls the latest image for the target deployment.

## Example Usage

```shell
terraform apply -invoke action.komodo_pull_deployment.example
```

```terraform
action "komodo_pull_deployment" "example" {
  config {
    deployment = "my-deployment"
  }
}
```

## Schema

### Config

#### Required

- `deployment` (String) Id or name of the deployment to pull.
