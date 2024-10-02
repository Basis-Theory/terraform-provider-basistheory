package main

import (
	"github.com/Basis-Theory/terraform-provider-basistheory/internal/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Format TF examples
//go:generate terraform fmt -recursive ./examples/

// Generate TF docs
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

var (
// these will be set by the goreleaser configuration
// to appropriate values for the compiled binary
// version string = "dev"

// goreleaser can also pass the specific commit if you want
// commit  string = ""
)

func main() {
	plugin.Serve(&plugin.ServeOpts{ProviderFunc: provider.BasisTheoryProvider(nil, nil)})
}
