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
