---
page_title: "v1 to v2 Migration - terraform-provider-basistheory"
subcategory: ""
description: |-
  How to migrate from v1 to v2 of the Basis Theory Terraform provider
---

# v1 to v2 Migration

This is a guide on how to upgrade from v1 to v2 of the Basis Theory Terraform provider. We try to automate as many steps
as possible for you, but there are a few manual steps that must be done on your end to ensure the upgrade is successful.
Before upgrading follow the steps for each resource in this guide.

## basistheory_reactor

If you currently have a `basistheory_reactor` with a `formula_id`, you will need to create a new reactor with the `code`
that you've set on the corresponding `basistheory_reactor_formula`. If this `basistheory_reactor` is currently being used in
production, you should:

1. Duplicate the `basistheory_reactor` with the same `code` as the corresponding `basistheory_reactor_formula`
2. Change your production systems to start using the newly created `basistheory_reactor`
3. Delete the old reactor

## basistheory_application `create_key` and basistheory_application_keys

In `v2` we have introduced the concept of [Application Keys](https://developers.basistheory.com/docs/api/applications/application-keys), which allow multiple keys to be created for an application - enabling 
you to rotate keys without downtime. With this addition, we've added a `create_key` property to the `basistheory_application` resource 
which is only used when creating a new application. 

- When creating a new `basistheory_application` in v2
  - We strongly suggest you leave the `create_key` default
  - Create a `basistheory_application_key` along with this, to get full lifecycle management of the key. 
- If you are migrating from a `v1` basistheory_application to `v2`
  - You will need to update your existing `basistheory_application` resources with the `create_key` property set to `true`
  - You are able to also create a `basistheory_application_key` for this application
  - ⚠️ Keep in mind that these applications will have an Unmanaged Key, which means you will need to manage the key lifecycle yourself in the portal.
  - If you'd like to fully migrate to the new key management system, you can:
    - Create a `basistheory_application_key` for this application 
    - Update your systems to use the new key
    - (VERIFY ALL LOCATIONS ARE UPDATED, THIS STEP CAN NOT BE REVERTED) Delete the old key from the Portal
    - You now have full lifecycle management of the Application Keys in Terraform