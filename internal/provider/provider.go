package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	basistheory "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/Basis-Theory/go-sdk/v5/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func BasisTheoryProvider(client *basistheory.Client) func() *schema.Provider {
	const BasisTheoryClientDefaultTimeout = 15

	return func() *schema.Provider {
		provider := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"api_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "API key for the BasisTheory client. Can be set through BASISTHEORY_API_KEY env var",
					DefaultFunc: schema.EnvDefaultFunc("BASISTHEORY_API_KEY", nil),
				},
				"api_url": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Base API URL for the BasisTheory client. Defaults to https://api.basistheory.com. Can be set through BASISTHEORY_API_URL env var",
					DefaultFunc: schema.EnvDefaultFunc("BASISTHEORY_API_URL", "https://api.basistheory.com"),
				},
				"client_timeout": {
					Optional:    true,
					Type:        schema.TypeInt,
					Description: "Timeout (in seconds) for the BasisTheory client. Defaults to 15 seconds. Can be set through BASISTHEORY_CLIENT_TIMEOUT env var",
					DefaultFunc: schema.EnvDefaultFunc("BASISTHEORY_CLIENT_TIMEOUT", BasisTheoryClientDefaultTimeout),
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"basistheory_applepay_domain":       resourceApplePayDomain(),
				"basistheory_reactor":               resourceBasisTheoryReactor(),
				"basistheory_application":           resourceBasisTheoryApplication(),
				"basistheory_proxy":                 resourceBasisTheoryProxy(),
				"basistheory_application_key":       resourceBasisTheoryApplicationKey(),
				"basistheory_webhook":               resourceBasisTheoryWebhook(),
				"basistheory_client_encryption_key": resourceBasisTheoryClientEncryptionKey(),
			},
		}
		provider.ConfigureContextFunc = configure(client, provider)

		return provider
	}
}

func configure(client *basistheory.Client, provider *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		if client != nil {
			return map[string]interface{}{
				"client":  client,
				"api_key": data.Get("api_key"),
				"api_url": data.Get("api_url"),
			}, nil
		}

		userAgent := fmt.Sprintf("HashiCorp Terraform/%s Basis Theory Terraform Plugin SDK/%s", provider.TerraformVersion, meta.SDKVersionString())

		var diags diag.Diagnostics

		return map[string]interface{}{
			"client":  newClient(data, userAgent),
			"api_key": data.Get("api_key"),
			"api_url": data.Get("api_url"),
		}, diags
	}
}

func newClient(data *schema.ResourceData, userAgent string) *basistheory.Client {
	return basistheory.NewClient(
		option.WithAPIKey(data.Get("api_key").(string)),
		option.WithBaseURL(data.Get("api_url").(string)),
		option.WithHTTPHeader(map[string][]string{
			"User-Agent": {userAgent},
		}),
		option.WithHTTPClient(
			&http.Client{
				Timeout: time.Duration(data.Get("client_timeout").(int)) * time.Second,
			},
		),
	)
}
