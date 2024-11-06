package provider

import (
	"context"
	"fmt"
	"github.com/Basis-Theory/basistheory-go/v6"
	basistheoryV2 "github.com/Basis-Theory/go-sdk"
	basistheoryV2client "github.com/Basis-Theory/go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryProxy() *schema.Resource {
	return &schema.Resource{
		Description: "Proxy https://docs.basistheory.com/docs/api/proxies/pre-configured-proxies",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceProxyCreate,
		ReadContext:   resourceProxyReadV2,
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
				Description:  "Request transform for the Proxy",
				Type:         schema.TypeMap,
				Optional:     true,
				ValidateFunc: validateRequestTransformProperties,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"response_transform": {
				Description:  "Response transform for the Proxy",
				Type:         schema.TypeMap,
				Optional:     true,
				ValidateFunc: validateResponseTransformProperties,
				Elem: &schema.Schema{
					Type: schema.TypeString,
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
	if proxy.RequestTransform != nil {
		proxyRequest.SetRequestTransform(proxy.GetRequestTransform())
	}
	if proxy.ResponseTransform != nil {
		proxyRequest.SetResponseTransform(proxy.GetResponseTransform())
	}
	proxyRequest.SetConfiguration(proxy.GetConfiguration())
	proxyRequest.SetRequireAuth(proxy.GetRequireAuth())

	application := *basistheory.NewApplication()
	applicationId := proxy.GetApplicationId()
	if applicationId != "" {
		application.SetId(applicationId)
		proxyRequest.SetApplication(application)
	}

	createdProxy, response, err := basisTheoryClient.ProxiesApi.Create(ctxWithApiKey).CreateProxyRequest(proxyRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Proxy:", response, err)
	}

	data.SetId(createdProxy.GetId())

	return resourceProxyReadV2(ctx, data, meta)
}

func resourceProxyReadV2(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	proxy, err := basisTheoryClient.Proxies.Get(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnosticsV2("Error reading Proxy:", err)
	}

	data.SetId(*proxy.ID)

	modifiedAt := ""

	if proxy.ModifiedAt != nil {
		modifiedAt = proxy.ModifiedAt.String()
	}

	for proxyDatumName, proxyDatum := range map[string]interface{}{
		"key":                 proxy.Key,
		"tenant_id":           proxy.TenantID,
		"name":                proxy.Name,
		"destination_url":     proxy.DestinationURL,
		"request_reactor_id":  proxy.RequestReactorID,
		"response_reactor_id": proxy.ResponseReactorID,
		"request_transform":   flattenRequestProxyTransformDataV2(proxy.RequestTransform),
		"response_transform":  flattenResponseProxyTransformDataV2(proxy.ResponseTransform),
		"application_id":      proxy.ApplicationID,
		"configuration":       proxy.Configuration,
		"require_auth":        proxy.RequireAuth,
		"created_at":          proxy.CreatedAt.String(),
		"created_by":          proxy.CreatedBy,
		"modified_at":         modifiedAt,
		"modified_by":         proxy.ModifiedBy,
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

	application := *basistheory.NewApplication()
	applicationId := proxy.GetApplicationId()
	if applicationId != "" {
		application.SetId(applicationId)
		updateProxyRequest.SetApplication(application)
	}

	_, response, err := basisTheoryClient.ProxiesApi.Update(ctxWithApiKey, proxy.GetId()).UpdateProxyRequest(updateProxyRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Proxy:", response, err)
	}

	return resourceProxyReadV2(ctx, data, meta)
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
	proxy.SetApplicationId(data.Get("application_id").(string))
	proxy.SetRequireAuth(data.Get("require_auth").(bool))

	if requestTransform, ok := data.Get("request_transform").(map[string]interface{}); ok {
		if requestTransform["code"] != nil {
			transform := *basistheory.NewProxyTransform()
			transform.SetCode(requestTransform["code"].(string))
			proxy.SetRequestTransform(transform)
		}
	}

	if responseTransform, ok := data.Get("response_transform").(map[string]interface{}); ok {
		if responseTransform["code"] != nil {
			transform := *basistheory.NewProxyTransform()
			transform.SetType("code")
			transform.SetCode(responseTransform["code"].(string))
			proxy.SetResponseTransform(transform)
		}

		if responseTransform["type"] != nil && responseTransform["type"].(string) == "mask" {
			transform := *basistheory.NewProxyTransform()
			transform.SetType("mask")
			transform.SetMatcher(responseTransform["matcher"].(string))
			if responseTransform["expression"] != nil {
				transform.SetExpression(responseTransform["expression"].(string))
			}
			transform.SetReplacement(responseTransform["replacement"].(string))
			proxy.SetResponseTransform(transform)
		}
	}

	configOptions := map[string]string{}
	for key, value := range data.Get("configuration").(map[string]interface{}) {
		configOptions[key] = value.(string)
	}

	proxy.SetConfiguration(configOptions)

	return proxy
}

func flattenRequestProxyTransformDataV2(proxyTransform *basistheoryV2.ProxyTransform) map[string]interface{} {
	transform := make(map[string]interface{})

	if proxyTransform != nil && proxyTransform.Code != nil {
		transform["code"] = proxyTransform.Code
	}

	return transform
}

func flattenResponseProxyTransformDataV2(proxyTransform *basistheoryV2.ProxyTransform) interface{} {
	transform := make(map[string]interface{})

	if proxyTransform != nil {
		if proxyTransform.Type != nil {
			transform["type"] = proxyTransform.Type
		}
		if proxyTransform.Code != nil {
			transform["code"] = proxyTransform.Code
		}
		if proxyTransform.Matcher != nil {
			transform["matcher"] = proxyTransform.Matcher
		}
		if proxyTransform.Expression != nil {
			transform["expression"] = proxyTransform.Expression
		}
		if proxyTransform.Replacement != nil {
			transform["replacement"] = proxyTransform.Replacement
		}
	}

	return transform
}

func validateRequestTransformProperties(val interface{}, _ string) (warns []string, errs []error) {
	transform := val.(map[string]interface{})
	if transform["code"] == nil || transform["code"] == "" {
		errs = append(errs, fmt.Errorf("code is required"))
	}

	for transformKey := range transform {
		if transformKey != "code" {
			errs = append(errs, fmt.Errorf("invalid transform property of: %s", transformKey))
		}
	}

	return
}

func validateResponseTransformProperties(val interface{}, _ string) (warns []string, errs []error) {
	transform := val.(map[string]interface{})
	allowedAttributes := map[string]struct{}{
		"type":        {},
		"code":        {},
		"matcher":     {},
		"replacement": {},
		"expression":  {},
	}

	for transformKey := range transform {
		if _, ok := allowedAttributes[transformKey]; !ok {
			errs = append(errs, fmt.Errorf("invalid transform property of: %s", transformKey))
		}
	}

	if len(errs) > 0 {
		return
	}

	if (!IsNilOrEmpty(transform["type"]) && transform["type"].(string) == "code") || !IsNilOrEmpty(transform["code"]) {
		if transform["type"].(string) != "code" {
			errs = append(errs, fmt.Errorf("type must be code when code is provided"))
		}
		if IsNilOrEmpty(transform["code"]) {
			errs = append(errs, fmt.Errorf("code is required when type is code"))
		}
		if !IsNilOrEmpty(transform["matcher"]) {
			errs = append(errs, fmt.Errorf("matcher is not valid when type is code"))
		}
		if !IsNilOrEmpty(transform["expression"]) {
			errs = append(errs, fmt.Errorf("expression is not valid when type is code"))
		}
		if !IsNilOrEmpty(transform["replacement"]) {
			errs = append(errs, fmt.Errorf("replacement is not valid when type is code"))
		}
	} else if !IsNilOrEmpty(transform["type"]) && transform["type"].(string) == "mask" {
		if IsNilOrEmpty(transform["matcher"]) {
			errs = append(errs, fmt.Errorf("matcher is required when type is mask"))
		}
		if IsNilOrEmpty(transform["replacement"]) {
			errs = append(errs, fmt.Errorf("replacement is required when type is mask"))
		}
		if !IsNilOrEmpty(transform["matcher"]) && transform["matcher"].(string) == "regex" && IsNilOrEmpty(transform["expression"]) {
			errs = append(errs, fmt.Errorf("expression is required when type is mask and matcher is regex"))
		}
	}

	return
}

func IsNilOrEmpty(value interface{}) bool {
	return value == nil || value == ""
}
