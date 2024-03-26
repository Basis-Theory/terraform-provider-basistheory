# Contributing

## Prerequisites

First ensure you have Go and Terraform installed:
```shell
brew install go terraform
```

## Updating dependencies

To update all dependencies to their latest versions, run the command:

```shell
go get -u
go mod tidy
```

To update a single dependency, say the `basistheory-go` SDK, you can run the command:

```shell
go get github.com/Basis-Theory/basistheory-go/v5
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
`/docs/resources` by running the command:

```shell
go generate
```
