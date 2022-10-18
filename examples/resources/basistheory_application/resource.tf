resource "basistheory_application" "my_example_application" {
  name = "My Example App"
  type = "private"
  permissions = [
    "token:create",
    "token:read",
    "token:use",
  ]
}

output "application_key" {
  value       = basistheory_application.my_example_application.key
  description = "My example application key"
  sensitive   = true
}
