resource "basistheory_application" "my_application" {
  name = "My App"
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
