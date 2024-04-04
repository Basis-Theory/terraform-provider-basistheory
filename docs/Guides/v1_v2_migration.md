---
page_title: "v1 to v2 Migration - terraform-provider-basistheory"
subcategory: ""
description: |-
  How to migrate from v1 to v2 of the Basis Theory Terraform provider
---

# v1 to v2 Migration

This is a guide on how to upgrade from v1 to v2 of the Basis Theory Terraform provider. For this upgrade, we try to automate
as many of these steps as possible for you. But there are a few manual steps that must be done on your end to ensure the
upgrade is successful. Before upgrading follow the steps for each resource in this guide.

## basistheory_reactor

If you currently have a `basistheory_reactor` with a `formula_id`, you will need to create a new reactor the `code` that
you've set on the corresponding `basistheory_reactor_formula`. If this `basistheory_reactor` is currently being used in
production, you should:

1. Duplicate the `basistheory_reactor` with the same `code` as the corresponding `basistheory_reactor_formula`
2. Change your production systems to start using the newly created `basistheory_reactor`
3. Delete the old reactor
