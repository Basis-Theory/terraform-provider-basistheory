resource "basistheory_reactor_formula" "my_private_reactor" {
  name        = "My Private Reactor"
  description = "Securely exchange token for another token"
  type        = "private"
  icon        = "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="
  code        = <<-EOT
  module.exports = async function (req) {
    // Do something with `req.configuration.SERVICE_API_KEY`

    return {
      raw: {
        foo: "bar"
      }
    };
  };
  EOT

  configuration {
    name        = "SERVICE_API_KEY"
    description = "Configuration description"
    type        = "string"
  }

  request_parameter {
    name        = "request_parameter_1"
    description = "Request parameter description"
    type        = "string"
  }

  request_parameter {
    name        = "request_parameter_2"
    description = "Request parameter description"
    type        = "boolean"
    optional    = true
  }
}
