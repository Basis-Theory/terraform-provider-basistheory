resource "basistheory_proxy" "my_proxy" {
  name            = "My Proxy"
  destination_url = "https://httpbin.org/post"
}

output "proxy_key" {
  value       = basistheory_proxy.my_proxy.key
  description = "My proxy key"
  sensitive   = true
}
