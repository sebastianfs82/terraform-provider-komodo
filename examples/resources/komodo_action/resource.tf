resource "komodo_action" "example" {
  name = "my-action"

  file_contents = <<-EOT
    export default async function run(): Promise<void> {
      console.log("Hello from Komodo action!");
    }
  EOT
}

resource "komodo_action" "scheduled" {
  name             = "nightly-cleanup"
  schedule_format  = "Cron"
  schedule         = "0 0 * * *"
  schedule_enabled = true
  failure_alert    = true
}
