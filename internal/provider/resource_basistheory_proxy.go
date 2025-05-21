package provider

import (
	"context"
	"fmt"
	basistheory "github.com/Basis-Theory/go-sdk/v2"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v2/client"
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
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	proxy := getProxyFromData(data)

	proxyRequest := &basistheory.CreateProxyRequest{
		Name: getStringValue(proxy.Name),
		DestinationURL: getStringValue(proxy.DestinationURL),
	}
	proxyRequest.RequestReactorID = proxy.RequestReactorID
	proxyRequest.ResponseReactorID = proxy.ResponseReactorID
	if proxy.RequestTransform != nil {
		proxyRequest.RequestTransform = proxy.RequestTransform
	}
	if proxy.ResponseTransform != nil {
		proxyRequest.ResponseTransform = proxy.ResponseTransform
	}
	proxyRequest.Configuration = proxy.Configuration
	proxyRequest.RequireAuth = proxy.RequireAuth

	application := &basistheory.Application {}
	applicationId := proxy.ApplicationID
	if applicationId != nil && *applicationId != "" {
		application.ID = applicationId
		proxyRequest.Application = application
	}

	createdProxy, err := basisTheoryClient.Proxies.Create(ctx, proxyRequest)

	if err != nil {
		return apiErrorDiagnostics("Error creating Proxy:", err)
	}

	data.SetId(*createdProxy.ID)

	return resourceProxyRead(ctx, data, meta)
}


func resourceProxyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	proxy, err := basisTheoryClient.Proxies.Get(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnostics("Error reading Proxy:", err)
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
		"request_transform":   flattenRequestProxyTransformData(proxy.RequestTransform),
		"response_transform":  flattenResponseProxyTransformData(proxy.ResponseTransform),
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
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	proxy := getProxyFromData(data)
	updateProxyRequest := &basistheory.UpdateProxyRequest{
		Name: getStringValue(proxy.Name),
		DestinationURL: getStringValue(proxy.DestinationURL),
	}
	updateProxyRequest.RequestReactorID = proxy.RequestReactorID
	updateProxyRequest.ResponseReactorID = proxy.ResponseReactorID
	updateProxyRequest.RequestTransform = proxy.RequestTransform
	updateProxyRequest.ResponseTransform = proxy.ResponseTransform
	updateProxyRequest.Configuration = proxy.Configuration
	updateProxyRequest.RequireAuth = proxy.RequireAuth

	application := &basistheory.Application {}
	applicationId := proxy.ApplicationID
	if applicationId != nil {
		application.ID = applicationId
		updateProxyRequest.Application = application
	}

	_, err := basisTheoryClient.Proxies.Update(ctx, getStringValue(proxy.ID), updateProxyRequest)

	if err != nil {
		return apiErrorDiagnostics("Error updating Proxy:", err)
	}

	return resourceProxyRead(ctx, data, meta)
}

func resourceProxyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	err := basisTheoryClient.Proxies.Delete(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnostics("Error deleting Proxy:", err)
	}

	return nil
}

func getProxyFromData(data *schema.ResourceData) basistheory.Proxy {
	id := data.Id()
	proxy := basistheory.Proxy{
		ID: &id,
		Name: getStringPointer(data.Get("name")),
		DestinationURL: getStringPointer(data.Get("destination_url")),
		RequestReactorID: getStringPointer(data.Get("request_reactor_id")),
		ResponseReactorID: getStringPointer(data.Get("response_reactor_id")),
		ApplicationID: getStringPointer(data.Get("application_id")),
		RequireAuth: getBoolPointer(data.Get("require_auth")),
	}

	if requestTransform, ok := data.Get("request_transform").(map[string]interface{}); ok {
		if requestTransform["code"] != nil && requestTransform["code"].(string) != "" {
			transform := &basistheory.ProxyTransform{
				Code: getStringPointer(requestTransform["code"]),
			}
			proxy.RequestTransform = transform
		}
	}

	if responseTransform, ok := data.Get("response_transform").(map[string]interface{}); ok {
		if responseTransform["code"] != nil && responseTransform["code"].(string) != "" {
			transform := &basistheory.ProxyTransform { }
			transform.Type = getStringPointer("code")
			transform.Code = getStringPointer(responseTransform["code"])
			proxy.ResponseTransform = transform
		}

		if responseTransform["type"] != nil && responseTransform["type"].(string) == "mask" {
			transform := &basistheory.ProxyTransform { }
			transform.Type = getStringPointer("mask")
			transform.Matcher = getStringPointer(responseTransform["matcher"])
			if responseTransform["expression"] != nil {
				transform.Expression = getStringPointer(responseTransform["expression"])
			}
			transform.Replacement = getStringPointer(responseTransform["replacement"])
			proxy.ResponseTransform = transform
		}
	}

	configOptions := map[string]*string{}
	for key, value := range data.Get("configuration").(map[string]interface{}) {
		configOptions[key] = getStringPointer(value)
	}

	proxy.Configuration = configOptions

	return proxy
}

func flattenRequestProxyTransformData(proxyTransform *basistheory.ProxyTransform) map[string]interface{} {
	transform := make(map[string]interface{})

	if proxyTransform != nil && proxyTransform.Code != nil {
		transform["code"] = proxyTransform.Code
	}

	return transform
}

func flattenResponseProxyTransformData(proxyTransform *basistheory.ProxyTransform) interface{} {
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
