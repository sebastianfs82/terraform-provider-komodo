data "komodo_version" "this" {}

output "komodo_version" {
  value = data.komodo_version.this.version
}
