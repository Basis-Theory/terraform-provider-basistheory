resource "basistheory_application" "my_example_application" {
  name = "My Example App"
  type = "private"
  rule {
    description = "Create and read masked tokens"
    priority    = 1
    container   = "/"
    transform   = "mask"
    permissions = [
      "token:create",
      "token:read",
    ]
  }
  rule {
    description = "Use plaintext tokens in services"
    priority    = 2
    container   = "/"
    transform   = "reveal"
    permissions = [
      "token:use",
    ]
  }
}

output "application_key" {
  value       = basistheory_application.my_example_application.key
  description = "My example application key"
  sensitive   = true
}
