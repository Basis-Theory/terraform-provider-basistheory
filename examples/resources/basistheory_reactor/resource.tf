resource "basistheory_reactor" "my_reactor" {
  name       = "My Reactor"
  formula_id = basistheory_reactor_formula.reactor_formula_resource_name.id
  configuration = {
    SERVICE_API_KEY = "key_abcd1234"
  }
}
