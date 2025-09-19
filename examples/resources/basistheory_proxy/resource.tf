resource "basistheory_proxy" "my_proxy" {
  name            = "My Proxy"
  destination_url = "https://httpbin.org/post"
}

# Proxy with Request Transforms - processes incoming requests
resource "basistheory_proxy" "request_transform_proxy" {
  name            = "Request Transform Proxy"
  destination_url = "https://api.example.com/payments"
  require_auth    = true

  # Request transforms - executed in order on incoming requests
  request_transforms {
    # Tokenize credit card data from request
    type = "tokenize"
    options = {
      identifier = "requestCardToken"
      token = jsonencode({
        type = "card"
        data = "{{ encrypted | json: '$.data' }}"
        metadata = {
          source = "proxy-request"
        }
      })
    }
  }
}

# Proxy with Response Transforms - processes outgoing responses
resource "basistheory_proxy" "response_transform_proxy" {
  name            = "Response Transform Proxy"
  destination_url = "https://api.bank.com/accounts"
  require_auth    = true

  # Response transforms - executed in order on outgoing responses
  response_transforms {
    # Tokenize sensitive account data from response
    type = "tokenize"
    options = {
      identifier = "responseAccountToken"
      token = jsonencode({
        type = "card"
        data = {
          "number" : "{{ res.number }}",
          "cvc" : "{{ res.cvc }}",
          "expiration_month" : "{{ res.expiration_month }}",
          "expiration_year" : "{{ res.expiration_year }}"
        },
        metadata = {
          source = "proxy-response"
        }
      })
    }
  }

  response_transforms {
    # Replace account number with token ID in response JSON
    type = "append_json"
    options = {
      value    = "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"
      location = "$.tokenized_account_id"
    }
  }

  response_transforms {
    # Add response header with token reference
    type = "append_header"
    options = {
      value    = "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"
      location = "X-Account-Token-ID"
    }
  }
}

output "proxy_key" {
  value       = basistheory_proxy.my_proxy.key
  description = "My proxy key"
  sensitive   = true
}

output "request_proxy_key" {
  value       = basistheory_proxy.request_transform_proxy.key
  description = "Request transform proxy key"
  sensitive   = true
}

output "response_proxy_key" {
  value       = basistheory_proxy.response_transform_proxy.key
  description = "Response transform proxy key"
  sensitive   = true
}
