resource "basistheory_reactor" "my_reactor" {
  name = "My Reactor"
  code = <<-EOT
  module.exports = async function (req) {
    // Do something with `req.configuration.SERVICE_API_KEY`

    return {
      raw: {
        foo: "bar"
      }
    };
  };
  EOT
  configuration = {
    SERVICE_API_KEY = "key_abcd1234"
  }
}

# Reactor on the node24 runtime, with an execution role and runtime logs
resource "basistheory_reactor" "my_node24_reactor" {
  name = "My Node24 Reactor"
  code = <<-EOT
  module.exports = async function (context) {
    return {
      res: {
        body: {},
        headers: {},
        statusCode: 200
      }
    };
  };
  EOT

  runtime {
    image     = "node24"
    timeout   = 10
    resources = "standard"

    # The access the code runs with: `permissions` or `rule` blocks
    execution_role {
      rule {
        description = "Reveal general tokens"
        priority    = 1
        container   = "/general/"
        transform   = "reveal"
        permissions = ["token:read"]
      }
    }

    # Without this block logs are off, and logs turned on outside Terraform
    # are turned off on the next apply
    logs {
      enabled = true
      level   = "info"
    }
  }
}
