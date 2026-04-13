# Look up a Server terminal by name
data "komodo_terminal" "server" {
  target_type = "Server"
  target_id   = "69db2f6e0816ddac8244a5b3"
  name        = "my-terminal"
}

# Look up a Container terminal by name
data "komodo_terminal" "container" {
  target_type = "Container"
  target_id   = "69db2f6e0816ddac8244a5b3" # the server hosting the container
  container   = "adguard"
  name        = "my-container-terminal"
}

# Look up a Stack service terminal by name
data "komodo_terminal" "stack" {
  target_type = "Stack"
  target_id   = "69dba72e0816ddac8244ae18"
  service     = "nginx"
  name        = "my-stack-terminal"
}

# Look up a Deployment terminal by name
data "komodo_terminal" "deployment" {
  target_type = "Deployment"
  target_id   = "69dba72e0816ddac8244ae18"
  name        = "my-deployment-terminal"
}
