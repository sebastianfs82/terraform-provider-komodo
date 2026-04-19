resource "komodo_deployment" "example" {
  name      = "my-app"
  server_id = komodo_server.example.id

  image {
    image = "nginx:latest"
  }

  lifecycle {
    action_trigger {
      events = [
        after_create,
        after_update,
      ]
      actions = [
        action.komodo_deployment.deploy
      ]
    }
  }
}

# Deploy (or redeploy) the container.
action "komodo_deployment" "deploy" {
  config {
    id          = komodo_deployment.example.id
    action      = "deploy"
    stop_signal = "SIGTERM"
    stop_time   = 30
  }
}

# Stop the container with a custom signal.
action "komodo_deployment" "stop" {
  config {
    id     = komodo_deployment.example.id
    action = "stop"
    signal = "SIGTERM"
    time   = 10
  }
}

# Destroy (stop + remove) the container.
action "komodo_deployment" "destroy" {
  config {
    id     = komodo_deployment.example.id
    action = "destroy"
  }
}

# Other available actions: start, restart, pull, pause, unpause
