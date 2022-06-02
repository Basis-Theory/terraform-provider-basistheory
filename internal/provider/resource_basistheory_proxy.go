package provider

import (
	"context"
	"github.com/Basis-Theory/basistheory-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryProxy() *schema.Resource {
	return &schema.Resource{
		Description: "Proxy", // TODO: add link if the new Proxy is documented

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceProxyCreate,
		ReadContext:   resourceProxyRead,
		UpdateContext: resourceProxyUpdate,
		DeleteContext: resourceProxyDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Proxy",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"key": {
				Description: "Key for the Proxy",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tenant_id": {
				Description: "Tenant identifier where this Proxy was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name for the Proxy",
				Type:        schema.TypeString,
				Required:    true,
			},
			"destination_url": {
				Description: "Destination URL for the Proxy",
				Type:        schema.TypeString,
				Required:    true,
			},
			"request_reactor_id": {
				Description: "Request reactor ID for the Proxy",
				Type:        schema.TypeString,
				Required:    true,
			},
			"created_at": {
				Description: "Timestamp at which the Proxy was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Proxy",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_at": {
				Description: "Timestamp at which the Proxy was last updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_by": {
				Description: "Identifier for who last modified the Proxy",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceProxyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	proxy := getProxyFromData(data)

	proxyRequest := *basistheory.NewCreateProxyRequest(proxy.GetName(), proxy.GetDestinationUrl(), proxy.GetRequestReactorId())

	createdProxy, response, err := basisTheoryClient.ProxiesApi.ProxiesCreate(ctxWithApiKey).CreateProxyRequest(proxyRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Proxy:", response, err)
	}

	data.SetId(createdProxy.GetId())

	return resourceProxyRead(ctx, data, meta)
}

func resourceProxyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	proxy, response, err := basisTheoryClient.ProxiesApi.ProxiesGetById(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading Proxy:", response, err)
	}

	data.SetId(proxy.GetId())

	modifiedAt := ""

	if proxy.ModifiedAt.IsSet() {
		modifiedAt = proxy.GetModifiedAt().String()
	}

	for proxyDatumName, proxyDatum := range map[string]interface{}{
		"key":                proxy.GetKey(),
		"tenant_id":          proxy.GetTenantId(),
		"name":               proxy.GetName(),
		"destination_url":    proxy.GetDestinationUrl(),
		"request_reactor_id": proxy.GetRequestReactorId(),
		"created_at":         proxy.GetCreatedAt().String(),
		"created_by":         proxy.GetCreatedBy(),
		"modified_at":        modifiedAt,
		"modified_by":        proxy.GetModifiedBy(),
	} {
		err := data.Set(proxyDatumName, proxyDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceProxyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	proxy := getProxyFromData(data)
	updateProxyRequest := *basistheory.NewUpdateProxyRequest(proxy.GetName(), proxy.GetDestinationUrl(), proxy.GetRequestReactorId())

	_, response, err := basisTheoryClient.ProxiesApi.ProxiesUpdate(ctxWithApiKey, proxy.GetId()).UpdateProxyRequest(updateProxyRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Proxy:", response, err)
	}

	return resourceProxyRead(ctx, data, meta)
}

func resourceProxyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	response, err := basisTheoryClient.ProxiesApi.ProxiesDelete(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error deleting Proxy:", response, err)
	}

	return nil
}

func getProxyFromData(data *schema.ResourceData) *basistheory.Proxy {
	proxy := basistheory.NewProxy()
	proxy.SetId(data.Id())
	proxy.SetName(data.Get("name").(string))
	proxy.SetDestinationUrl(data.Get("destination_url").(string))
	proxy.SetRequestReactorId(data.Get("request_reactor_id").(string))

	return proxy
}
