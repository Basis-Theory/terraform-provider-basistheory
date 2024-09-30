package provider

import (
	"context"
	"fmt"
	"github.com/Basis-Theory/basistheory-go/v6"
	basistheoryV2 "github.com/Basis-Theory/go-sdk/client"
	"github.com/Basis-Theory/go-sdk/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func BasisTheoryProvider(client *basistheory.APIClient, clientV2 *basistheoryV2.Client) func() *schema.Provider {
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
				"basistheory_reactor":         resourceBasisTheoryReactor(),
				"basistheory_application":     resourceBasisTheoryApplication(),
				"basistheory_proxy":           resourceBasisTheoryProxy(),
				"basistheory_application_key": resourceBasisTheoryApplicationKey(),
				"basistheory_webhook":         resourceBasisTheoryWebhook(),
			},
		}

		provider.ConfigureContextFunc = configure(client, clientV2, provider)

		return provider
	}
}

func configure(client *basistheory.APIClient, clientV2 *basistheoryV2.Client, provider *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
		if client != nil && clientV2 != nil {
			return map[string]interface{}{
				"client":  client,
				"clientV2": clientV2,
				"api_key": data.Get("api_key"),
			}, nil
		}

		url := data.Get("api_url").(string)
		clientTimeout := data.Get("client_timeout").(int)

		userAgent := fmt.Sprintf("HashiCorp Terraform/%s Basis Theory Terraform Plugin SDK/%s", provider.TerraformVersion, meta.SDKVersionString())

		var diags diag.Diagnostics

		return map[string]interface{}{
			"client":   newClientV1(url, userAgent, clientTimeout),
			"clientV2": newClientV2(data, userAgent),
			"api_key":  data.Get("api_key"),
		}, diags
	}
}

func newClientV1(url string, userAgent string, clientTimeout int) *basistheory.APIClient {
	urlArray := strings.Split(url, "://")
	configuration := basistheory.NewConfiguration()
	configuration.Scheme = urlArray[0]
	configuration.Host = urlArray[1]
	configuration.UserAgent = userAgent
	configuration.DefaultHeader = map[string]string{
		"Keep-Alive": strconv.Itoa(clientTimeout),
	}
	apiClient := basistheory.NewAPIClient(configuration)
	return apiClient
}

func newClientV2(data *schema.ResourceData, userAgent string) *basistheoryV2.Client {
	return basistheoryV2.NewClient(
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
