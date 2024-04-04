resource "basistheory_application_key" "my_application_key" {
  application_id = basistheory_application.my_application.id
}

output "application_key" {
  value       = basistheory_application_key.my_application_key.key
  description = "My application key"
  sensitive   = true
}
