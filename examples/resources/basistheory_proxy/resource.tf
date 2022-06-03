resource "basistheory_proxy" "my_proxy" {
  name               = "My Proxy"
  destination_url    = "https://httpbin.org/post"
  request_reactor_id = basistheory_reactor.reactor_resource_name.id
  require_auth       = true
}

output "proxy_key" {
  value       = basistheory_proxy.my_proxy.key
  description = "My proxy key"
  sensitive   = true
}
