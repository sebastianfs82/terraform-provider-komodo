resource "komodo_stack" "example" {
  name      = "my-stack"
  server_id = komodo_server.example.id

  compose {
    contents = file("${path.module}/compose.yaml")
  }
}

# Nginx stack with inline compose contents
resource "komodo_stack" "nginx" {
  name      = "nginx"
  server_id = komodo_server.example.id

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
}

# Stack sourced from a git repository
resource "komodo_provider_account" "example" {
  domain   = "github.com"
  username = "myuser"
  token    = var.github_token
}

resource "komodo_stack" "from_git" {
  name      = "my-git-stack"
  server_id = komodo_server.example.id

  source {
    path       = "myorg/my-stack-repo"
    branch     = "main"
    account_id = komodo_provider_account.example.id
  }
}
