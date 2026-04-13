# Terminal on a Server
resource "komodo_terminal" "server" {
  name        = "my-terminal"
  target_type = "Server"
  target_id   = "69db2f6e0816ddac8244a5b3" # the server id
  command     = "bash"
}

# Terminal inside a Container using exec mode (default)
resource "komodo_terminal" "container_exec" {
  name        = "my-container-terminal"
  target_type = "Container"
  target_id   = "69db2f6e0816ddac8244a5b3" # the server hosting the container
  container   = "adguard"
  mode        = "exec"
  command     = "sh"
}

# Terminal inside a Container using attach mode
resource "komodo_terminal" "container_attach" {
  name        = "my-container-attach"
  target_type = "Container"
  target_id   = "69db2f6e0816ddac8244a5b3" # the server hosting the container
  container   = "adguard"
  mode        = "attach"
}

# Terminal inside a Stack service using exec mode (default)
resource "komodo_terminal" "stack_exec" {
  name        = "my-stack-terminal"
  target_type = "Stack"
  target_id   = "69dba72e0816ddac8244ae18" # the stack id
  service     = "nginx"
  mode        = "exec"
  command     = "sh"
}

# Terminal inside a Stack service using attach mode
resource "komodo_terminal" "stack_attach" {
  name        = "my-stack-attach"
  target_type = "Stack"
  target_id   = "69dba72e0816ddac8244ae18" # the stack id
  service     = "nginx"
  mode        = "attach"
}

# Terminal inside a Deployment
resource "komodo_terminal" "deployment" {
  name        = "my-deployment-terminal"
  target_type = "Deployment"
  target_id   = "69dba72e0816ddac8244ae18" # the deployment id
  command     = "sh"
}
