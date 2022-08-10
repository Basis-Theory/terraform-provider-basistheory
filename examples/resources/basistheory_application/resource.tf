resource "basistheory_application" "my_example_application" {
  name = "My Example App"
  type = "private"
  permissions = [
    "token:general:create",
    "token:general:read:low",
    "token:pci:create",
    "token:pci:read:low",
  ]
}

output "application_key" {
  value       = basistheory_application.my_example_application.key
  description = "My example application key"
  sensitive   = true
}
