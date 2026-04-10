---
page_title: "komodo_repo_pull Action - terraform-provider-komodo"
subcategory: ""
description: |-
  Pulls the latest changes for the target repo on its attached server.
---

# komodo_repo_pull (Action)

Pulls the latest changes for the target repo on its attached server.

The repo must have a server attached (`server_id`). Komodo will run `git pull`
on the target server, then execute any `on_pull` hook defined on the repo.

## Example Usage

Actions are not executed automatically — they must be triggered either by an
`action_trigger` block in a resource lifecycle, or invoked explicitly on the CLI:

```shell
terraform apply -invoke action.komodo_repo_pull.example
```

### Trigger after repo resource changes

```terraform
resource "komodo_repo" "my_repo" {
  name = "my-repo"
  # ...
}

action "komodo_repo_pull" "example" {
  config {
    repo = komodo_repo.my_repo.name
  }
}

# Automatically pull whenever the repo resource is updated.
resource "terraform_data" "pull_trigger" {
  input = komodo_repo.my_repo.id

  lifecycle {
    action_trigger {
      events  = [after_update]
      actions = [action.komodo_repo_pull.example]
    }
  }
}
```

## Schema

### Config

#### Required

- `repo` (String) Id or name of the repo to pull.
