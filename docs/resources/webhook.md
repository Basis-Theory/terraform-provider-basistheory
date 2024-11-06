---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "basistheory_webhook Resource - terraform-provider-basistheory"
subcategory: ""
description: |-
  
---

# basistheory_webhook (Resource)





<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `events` (Set of String) List of events to subscribe to the Webhook
- `name` (String) Name of the Webhook
- `url` (String) URL of the Webhook

### Read-Only

- `created_at` (String) Timestamp at which the Webhook was created
- `created_by` (String) Identifier for who created the Webhook
- `id` (String) Unique identifier of the Webhook
- `modified_at` (String) Timestamp at which the Webhook was last updated
- `modified_by` (String) Identifier for who last modified the Webhook
- `tenant_id` (String) Tenant identifier where this Webhook was created

