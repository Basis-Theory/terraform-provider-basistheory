# Contributing

## Prerequisites

First ensure you have Go and Terraform installed:
```shell
brew install go terraform
```

## Running tests

Copy the `.env.example` file to `.env.local` and enter a valid Basis Theory 
management API key into this file.

To run tests using this configuration, run:

```shell
make verify
```

## Updating examples

The examples included under `/examples/resources` should be manually updated
with any new resources that are introduced.

After updating the examples, reformat and regenerate the markdown docs under 
`/docs/resources` by running the commands:

```shell
terraform fmt -recursive ./examples/
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
```
