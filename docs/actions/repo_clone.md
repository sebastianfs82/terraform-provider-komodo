---
page_title: "komodo_repo_clone Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Clones the target repo onto its attached server.
---

# komodo_repo_clone (Action)

Clones the target repo onto its attached server.

The repo must have a server attached (`server_id`). Komodo will run
`git clone` on the target server, then execute any `on_clone`/`on_pull` hooks
defined on the repo.

## Example Usage

Actions are not executed automatically — they must be triggered either by an
`action_trigger` block in a resource lifecycle, or invoked explicitly on the CLI:

```shell
terraform apply -invoke action.komodo_repo_clone.example
```

### Trigger after repo resource changes

```terraform
resource "komodo_repo" "my_repo" {
  name = "my-repo"
  # ...
}

action "komodo_repo_clone" "example" {
  config {
    repo = komodo_repo.my_repo.name
  }
}

# Automatically clone whenever the repo resource is created.
resource "terraform_data" "clone_trigger" {
  input = komodo_repo.my_repo.id

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.komodo_repo_clone.example]
    }
  }
}
```

## Schema

### Config

#### Required

- `repo` (String) Id or name of the repo to clone.
