---
page_title: "v3 to v4 Migration - terraform-provider-basistheory"
subcategory: ""
description: |-
  How to migrate from v3 to v4 of the Basis Theory Terraform provider
---

# v3 to v4 Migration

This guide explains how to upgrade from v3 to v4 of the Basis Theory Terraform provider. The primary breaking change in v4 affects the basistheory_proxy resource transform options syntax.

In v3, the `options` attribute for both `request_transforms` and `response_transforms` was expressed as a map (e.g., `options = { ... }`). In v4, `options` is now a proper nested block with named fields. This change improves validation, readability, and plan diffs.

Below are details and concrete examples to help you migrate.

## Affected resource
- basistheory_proxy
  - request_transforms[].options
  - response_transforms[].options

## What changed
- v3: `options` was a map attribute (key/value pairs), e.g. `options = { identifier = "...", value = "..." }`.
- v4: `options` is a nested block with explicit attributes, e.g.
  ```hcl
  request_transforms {
    type = "tokenize"
    options {
      identifier = "requestCardToken"
      token      = jsonencode({
        type = "card"
        data = "{{ encrypted | json: '$.data' }}"
      })
    }
  }
  ```

Supported fields under the new `options` block:
- identifier (String)
- value (String)
- location (String)
- token (String; typically a jsonencode(...) string)

## How to migrate your HCL
Convert each key/value inside the old `options = { ... }` map into attributes inside the `options { ... }` block.

- identifier = ... → options { identifier = "..." }
- value = ... → options { value = "..." }
- location = ... → options { location = "..." }
- token = ... → options { token = "..." } (if you used jsonencode(...) before, keep using it as the token string)

If you previously supplied `token` as a raw JSON string, we recommend switching to `jsonencode({ ... })` to simplify quoting and enable JSON-insensitive diffs.

### Request transforms — tokenize (before → after)
- v3 (before)
  ```hcl
  request_transforms {
    type = "tokenize"
    options = {
      identifier = "requestCardToken"
      token = jsonencode({
        type = "card"
        data = "{{ encrypted | json: '$.data' }}"
      })
    }
  }
  ```
- v4 (after)
  ```hcl
  request_transforms {
    type = "tokenize"
    options {
      identifier = "requestCardToken"
      token = jsonencode({
        type = "card"
        data = "{{ encrypted | json: '$.data' }}"
      })
    }
  }
  ```

### Response transforms — append_json (before → after)
- v3 (before)
  ```hcl
  response_transforms {
    type = "append_json"
    options = {
      value    = "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"
      location = "$.tokenized_account_id"
    }
  }
  ```
- v4 (after)
  ```hcl
  response_transforms {
    type = "append_json"
    options {
      value    = "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"
      location = "$.tokenized_account_id"
    }
  }
  ```

### Response transforms — append_header (before → after)
- v3 (before)
  ```hcl
  response_transforms {
    type = "append_header"
    options = {
      value    = "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"
      location = "X-Account-Token-ID"
    }
  }
  ```
- v4 (after)
  ```hcl
  response_transforms {
    type = "append_header"
    options {
      value    = "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"
      location = "X-Account-Token-ID"
    }
  }
  ```


Notes:
- v4 enforces types for each `options` attribute, providing clearer diffs and validation.
- For the `token` field, the provider performs JSON-insensitive diffs so formatting changes (whitespace, object key order) won’t cause unnecessary updates.

## Plan/apply expectations
- After updating your HCL to the new block syntax, run `terraform plan` to verify the changes. You should not need to recreate proxies; this is a configuration-only change.
- If Terraform shows a replacement for a proxy due solely to formatting of `token`, consider wrapping token JSON with `jsonencode(...)` as shown above.

## Troubleshooting
- Unknown argument errors inside `options`: Ensure you are using v4 of the provider and that `options` is declared as a block, not a map.
- Invalid JSON for `token`: Make sure you provide a string. Using `jsonencode({ ... })` is recommended.

## Version pinning
Update your required provider version to v4:
```hcl
terraform {
  required_providers {
    basistheory = {
      source  = "Basis-Theory/basistheory"
      version = ">= 4.0.0"
    }
  }
}
```
