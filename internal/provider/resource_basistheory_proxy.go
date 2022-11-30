package provider

import (
	"context"

	"github.com/Basis-Theory/basistheory-go/v3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryProxy() *schema.Resource {
	return &schema.Resource{
		Description: "Proxy https://docs.basistheory.com/#proxies",

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
				Sensitive:   true,
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
				Optional:    true,
				Default:     "",
			},
			"response_reactor_id": {
				Description: "Response reactor ID for the Proxy",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"request_transform": {
				Description: "Request transform for the Proxy",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Description: "The code that is executed when the Transform runs.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"response_transform": {
				Description: "Response transform for the Proxy",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"code": {
							Description: "The code that is executed when the Transform runs.",
							Type:        schema.TypeString,
							Required:    true,
						},
					},
				},
			},
			"configuration": {
				Description: "Configuration for the Reactor",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"application_id": {
				Description: "The Application's API key used in the BasisTheory instance passed into the Proxy Transform",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"require_auth": {
				Description: "Require auth for the Proxy",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
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

	proxyRequest := *basistheory.NewCreateProxyRequest(proxy.GetName(), proxy.GetDestinationUrl())
	proxyRequest.SetRequestReactorId(proxy.GetRequestReactorId())
	proxyRequest.SetResponseReactorId(proxy.GetResponseReactorId())
	proxyRequest.SetRequestTransform(proxy.GetRequestTransform())
	proxyRequest.SetResponseTransform(proxy.GetResponseTransform())
	proxyRequest.SetConfiguration(proxy.GetConfiguration())
	proxyRequest.SetRequireAuth(proxy.GetRequireAuth())

	if application, ok := proxy.GetApplicationOk(); ok {
		proxyRequest.SetApplication(*application)
	}

	createdProxy, response, err := basisTheoryClient.ProxiesApi.Create(ctxWithApiKey).CreateProxyRequest(proxyRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Proxy:", response, err)
	}

	data.SetId(createdProxy.GetId())

	return resourceProxyRead(ctx, data, meta)
}

func resourceProxyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	proxy, response, err := basisTheoryClient.ProxiesApi.GetById(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading Proxy:", response, err)
	}

	data.SetId(proxy.GetId())

	modifiedAt := ""

	if proxy.ModifiedAt.IsSet() {
		modifiedAt = proxy.GetModifiedAt().String()
	}

	for proxyDatumName, proxyDatum := range map[string]interface{}{
		"key":                 proxy.GetKey(),
		"tenant_id":           proxy.GetTenantId(),
		"name":                proxy.GetName(),
		"destination_url":     proxy.GetDestinationUrl(),
		"request_reactor_id":  proxy.GetRequestReactorId(),
		"response_reactor_id": proxy.GetResponseReactorId(),
		"request_transform":   proxy.GetRequestTransform(),
		"response_transform":  proxy.GetResponseTransform(),
		"application_id":      proxy.GetApplicationId(),
		"configuration":       proxy.GetConfiguration(),
		"require_auth":        proxy.GetRequireAuth(),
		"created_at":          proxy.GetCreatedAt().String(),
		"created_by":          proxy.GetCreatedBy(),
		"modified_at":         modifiedAt,
		"modified_by":         proxy.GetModifiedBy(),
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
	updateProxyRequest := *basistheory.NewUpdateProxyRequest(proxy.GetName(), proxy.GetDestinationUrl())
	updateProxyRequest.SetRequestReactorId(proxy.GetRequestReactorId())
	updateProxyRequest.SetResponseReactorId(proxy.GetResponseReactorId())
	updateProxyRequest.SetRequestTransform(proxy.GetRequestTransform())
	updateProxyRequest.SetResponseTransform(proxy.GetResponseTransform())
	updateProxyRequest.SetConfiguration(proxy.GetConfiguration())
	updateProxyRequest.SetRequireAuth(proxy.GetRequireAuth())

	if application, ok := proxy.GetApplicationOk(); ok {
		updateProxyRequest.SetApplication(*application)
	}

	_, response, err := basisTheoryClient.ProxiesApi.Update(ctxWithApiKey, proxy.GetId()).UpdateProxyRequest(updateProxyRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Proxy:", response, err)
	}

	return resourceProxyRead(ctx, data, meta)
}

func resourceProxyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	response, err := basisTheoryClient.ProxiesApi.Delete(ctxWithApiKey, data.Id()).Execute()

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
	proxy.SetResponseReactorId(data.Get("response_reactor_id").(string))
	proxy.SetRequireAuth(data.Get("require_auth").(bool))

	return proxy
}
