resource "komodo_action" "example" {
  name = "my-action"

  file_contents = <<-EOT
    export default async function run(): Promise<void> {
      console.log("Hello from Komodo action!");
    }
  EOT
}

resource "komodo_action" "scheduled" {
  name                  = "nightly-cleanup"
  failure_alert_enabled = true
  schedule {
    format     = "Cron"
    expression = "0 0 * * *"
    enabled    = true
  }
}
