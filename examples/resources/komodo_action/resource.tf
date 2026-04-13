resource "komodo_action" "example" {
  name = "my-action"

  file_contents = "console.log(\"Hello from Komodo action!\");"

  schedule {
    format     = "Cron"
    expression = "0 0 0 * * *"
    enabled    = true
  }

  argument {
    name  = "ENVIRONMENT"
    value = "production"
  }

  argument {
    name  = "LOG_LEVEL"
    value = "info"
  }

  argument {
    name  = "TIMEOUT"
    value = "30"
  }
}
