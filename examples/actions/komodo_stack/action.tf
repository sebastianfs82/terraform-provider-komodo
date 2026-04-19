resource "komodo_stack" "example" {
  name      = "nginx"
  server_id = komodo_servers.main.example.id

  compose {
    contents = <<-EOT
      services:
        nginx:
          image: nginx:latest
          ports:
            - "80:80"
          restart: unless-stopped
    EOT
  }

  lifecycle {
    action_trigger {
      events = [
        after_create,
        after_update,
      ]
      actions = [
        action.komodo_stack.deploy
      ]
    }
  }
}

# Deploy the stack whenever it is created or updated.
action "komodo_stack" "deploy" {
  config {
    id     = komodo_stack.example.id
    action = "deploy"
  }
}

# Destroy a specific service within the stack on-demand.
action "komodo_stack" "destroy" {
  config {
    id             = komodo_stack.example.id
    action         = "destroy"
    services       = ["nginx"]
    remove_orphans = true
  }
}

# Run a one-off command inside the nginx service container.
action "komodo_stack" "run_service" {
  config {
    id      = komodo_stack.example.id
    action  = "run_service"
    service = "nginx"
    command = ["nginx", "-t"]
    detach  = false
  }
}
