
package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	basistheory "github.com/Basis-Theory/go-sdk/v4"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v4/client"
	"github.com/Basis-Theory/go-sdk/v4/option"
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

		SchemaVersion: 1, // Increment schema version for the migration
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceBasisTheoryProxyResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceBasisTheoryProxyStateUpgradeV0,
				Version: 0,
			},
		},

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
			"state": {
				Description: "Current state of the Proxy",
				Type:        schema.TypeString,
				Computed:    true,
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
			"encrypted": {
				Description: "Base64-encoded encrypted token request data",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
			},
			"request_transforms": {
				Description: "Request transforms for the Proxy",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {Type: schema.TypeString, Optional: true},
						"code": {Type: schema.TypeString, Optional: true},
						"matcher": {Type: schema.TypeString, Optional: true},
						"expression": {Type: schema.TypeString, Optional: true},
						"replacement": {Type: schema.TypeString, Optional: true},

						"options": {
							Description: "Options for tokenize, append, and code transforms",
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// Existing tokenize/append fields
									"identifier": {
										Description: "Identifier for tokenize transforms",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"value": {
										Description: "Value for append transforms",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"location": {
										Description: "Location for append transforms",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"token": {
										Description: "Token configuration for tokenize transforms (JSON string)",
										Type:        schema.TypeString,
										Optional:    true,
									},
									// New runtime block
									"runtime": {
										Description: "Runtime configuration for code transforms",
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"image": {Type: schema.TypeString, Optional: true},
												"dependencies": {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
												"warm_concurrency": {Type: schema.TypeInt, Optional: true},
												"timeout": {Type: schema.TypeInt, Optional: true},
												"resources": {Type: schema.TypeString, Optional: true},
												"permissions": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
											},
										},
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// For token field, suppress diff if JSON content is equivalent
								if strings.HasSuffix(k, ".token") {
									return jsonEqual(old, new)
								}
								return old == new
							},
						},
					},
				},
			},
			"response_transforms": {
				Description: "Response transforms for the Proxy",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"code": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"matcher": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"expression": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"replacement": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"options": {
							Description: "Options for tokenize, append, and code transforms",
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// Existing tokenize/append fields
									"identifier": {
										Description: "Identifier for tokenize transforms",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"value": {
										Description: "Value for append transforms",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"location": {
										Description: "Location for append transforms",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"token": {
										Description: "Token configuration for tokenize transforms (JSON string)",
										Type:        schema.TypeString,
										Optional:    true,
									},
									// New runtime block
									"runtime": {
										Description: "Runtime configuration for code transforms",
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"image": {Type: schema.TypeString, Optional: true},
												"dependencies": {Type: schema.TypeMap, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
												"warm_concurrency": {Type: schema.TypeInt, Optional: true},
												"timeout": {Type: schema.TypeInt, Optional: true},
												"resources": {Type: schema.TypeString, Optional: true},
												"permissions": {Type: schema.TypeList, Optional: true, Elem: &schema.Schema{Type: schema.TypeString}},
											},
										},
									},
								},
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// For token field, suppress diff if JSON content is equivalent
								if strings.HasSuffix(k, ".token") {
									return jsonEqual(old, new)
								}
								return old == new
							},
						},
					},
				},
			},
		},
	}
}

// resourceBasisTheoryProxyResourceV0 defines the old schema (version 0) for migration
func resourceBasisTheoryProxyResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"tenant_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"destination_url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"configuration": {
				Type: schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"application_id": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"require_auth": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encrypted": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"request_transforms": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type":        {Type: schema.TypeString, Optional: true},
						"code":        {Type: schema.TypeString, Optional: true},
						"matcher":     {Type: schema.TypeString, Optional: true},
						"expression":  {Type: schema.TypeString, Optional: true},
						"replacement": {Type: schema.TypeString, Optional: true},
						// OLD: options was a map in v0
						"options": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"response_transforms": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type":        {Type: schema.TypeString, Optional: true},
						"code":        {Type: schema.TypeString, Optional: true},
						"matcher":     {Type: schema.TypeString, Optional: true},
						"expression":  {Type: schema.TypeString, Optional: true},
						"replacement": {Type: schema.TypeString, Optional: true},
						// OLD: options was a map in v0
						"options": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

// resourceBasisTheoryProxyStateUpgradeV0 migrates state from v0 to v1
func resourceBasisTheoryProxyStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	// Migrate request_transforms
	if requestTransforms, ok := rawState["request_transforms"].([]interface{}); ok {
		for i, transform := range requestTransforms {
			if transformMap, ok := transform.(map[string]interface{}); ok {
				if optionsMap, ok := transformMap["options"].(map[string]interface{}); ok && len(optionsMap) > 0 {
					// Convert map to list with single item
					transformMap["options"] = []interface{}{optionsMap}
					requestTransforms[i] = transformMap
				}
			}
		}
		rawState["request_transforms"] = requestTransforms
	}

	// Migrate response_transforms
	if responseTransforms, ok := rawState["response_transforms"].([]interface{}); ok {
		for i, transform := range responseTransforms {
			if transformMap, ok := transform.(map[string]interface{}); ok {
				if optionsMap, ok := transformMap["options"].(map[string]interface{}); ok && len(optionsMap) > 0 {
					// Convert map to list with single item
					transformMap["options"] = []interface{}{optionsMap}
					responseTransforms[i] = transformMap
				}
			}
		}
		rawState["response_transforms"] = responseTransforms
	}

	return rawState, nil
}

func resourceProxyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	// Validate transforms
	if requestTransforms, ok := data.GetOk("request_transforms"); ok {
		_, errs := validateRequestTransforms(requestTransforms, "request_transforms")
		if len(errs) > 0 {
			var diagErrors diag.Diagnostics
			for _, err := range errs {
				diagErrors = append(diagErrors, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid request_transforms configuration",
					Detail:   err.Error(),
				})
			}
			return diagErrors
		}
	}

	if responseTransforms, ok := data.GetOk("response_transforms"); ok {
		_, errs := validateResponseTransforms(responseTransforms, "response_transforms")
		if len(errs) > 0 {
			var diagErrors diag.Diagnostics
			for _, err := range errs {
				diagErrors = append(diagErrors, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid response_transforms configuration",
					Detail:   err.Error(),
				})
			}
			return diagErrors
		}
	}

	proxy := getProxyFromData(data)

	proxyRequest := &basistheory.CreateProxyRequest{
		Name:           getStringValue(proxy.Name),
		DestinationURL: getStringValue(proxy.DestinationURL),
	}
	if proxy.RequestTransforms != nil {
		proxyRequest.RequestTransforms = proxy.RequestTransforms
	}
	if proxy.ResponseTransforms != nil {
		proxyRequest.ResponseTransforms = proxy.ResponseTransforms
	}
	proxyRequest.Configuration = proxy.Configuration
	proxyRequest.RequireAuth = proxy.RequireAuth

	application := &basistheory.Application{}
	applicationId := proxy.ApplicationID
	if applicationId != nil && *applicationId != "" {
		application.ID = applicationId
		proxyRequest.Application = application
	}

	var requestOptions []option.IdempotentRequestOption
	if encryptedData, ok := data.GetOk("encrypted"); ok {
		if encryptedStr, ok := encryptedData.(string); ok && encryptedStr != "" {
			requestOptions = append(requestOptions, option.WithHTTPHeader(map[string][]string{
				"BT-ENCRYPTED": {encryptedStr},
			}))
		}
	}

	var createdProxy *basistheory.Proxy
	var err error

	if len(requestOptions) > 0 {
		createdProxy, err = basisTheoryClient.Proxies.Create(ctx, proxyRequest, requestOptions...)
	} else {
		createdProxy, err = basisTheoryClient.Proxies.Create(ctx, proxyRequest)
	}

	if err != nil {
		return apiErrorDiagnostics("Error creating Proxy:", err)
	}

	data.SetId(*createdProxy.ID)

	// Wait for the proxy to reach a final state before returning
	if diags := waitForProxyFinalState(ctx, basisTheoryClient, data.Id()); diags != nil {
		return diags
	}

	return resourceProxyRead(ctx, data, meta)
}

func waitForProxyFinalState(ctx context.Context, client *basistheoryClient.Client, id string) diag.Diagnostics {
	// Poll every 2 seconds up to 10 minutes
	interval := 2 * time.Second
	deadline := time.Now().Add(10 * time.Minute)

	for {
		if time.Now().After(deadline) {
			return diag.Errorf("timeout waiting for proxy %s to reach a final state", id)
		}

		proxy, err := client.Proxies.Get(ctx, id)
		if err != nil {
			return apiErrorDiagnostics("Error polling Proxy:", err)
		}

		state := ""
		if proxy.State != nil {
			state = *proxy.State
		}

		switch state {
		case "active":
			return nil
		case "failed", "outdated":
			return diag.Errorf("proxy %s reached failed state", id)
		}

		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err())
		case <-time.After(interval):
		}
	}
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

	// Set basic attributes
	basicAttributes := map[string]interface{}{
		"key":             proxy.Key,
		"tenant_id":       proxy.TenantID,
		"name":            proxy.Name,
		"destination_url": proxy.DestinationURL,
		"application_id":  proxy.ApplicationID,
		"configuration":   proxy.Configuration,
		"require_auth":    proxy.RequireAuth,
		"state":           proxy.State,
		"created_at":      proxy.CreatedAt.String(),
		"created_by":      proxy.CreatedBy,
		"modified_at":     modifiedAt,
		"modified_by":     proxy.ModifiedBy,
	}

	for proxyDatumName, proxyDatum := range basicAttributes {
		err := data.Set(proxyDatumName, proxyDatum)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle transforms
	err = data.Set("request_transforms", flattenProxyTransforms(proxy.RequestTransforms))
	if err != nil {
		return diag.FromErr(err)
	}

	err = data.Set("response_transforms", flattenProxyTransforms(proxy.ResponseTransforms))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceProxyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	// Validate transforms
	if requestTransforms, ok := data.GetOk("request_transforms"); ok {
		_, errs := validateRequestTransforms(requestTransforms, "request_transforms")
		if len(errs) > 0 {
			var diagErrors diag.Diagnostics
			for _, err := range errs {
				diagErrors = append(diagErrors, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid request_transforms configuration",
					Detail:   err.Error(),
				})
			}
			return diagErrors
		}
	}

	if responseTransforms, ok := data.GetOk("response_transforms"); ok {
		_, errs := validateResponseTransforms(responseTransforms, "response_transforms")
		if len(errs) > 0 {
			var diagErrors diag.Diagnostics
			for _, err := range errs {
				diagErrors = append(diagErrors, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid response_transforms configuration",
					Detail:   err.Error(),
				})
			}
			return diagErrors
		}
	}

	proxy := getProxyFromData(data)
	updateProxyRequest := &basistheory.UpdateProxyRequest{
		Name:           getStringValue(proxy.Name),
		DestinationURL: getStringValue(proxy.DestinationURL),
	}
	updateProxyRequest.RequestTransforms = proxy.RequestTransforms
	updateProxyRequest.ResponseTransforms = proxy.ResponseTransforms
	updateProxyRequest.Configuration = proxy.Configuration
	updateProxyRequest.RequireAuth = proxy.RequireAuth

	application := &basistheory.Application{}
	applicationId := proxy.ApplicationID
	if applicationId != nil && *applicationId != "" {
		application.ID = applicationId
		updateProxyRequest.Application = application
	} else {
		updateProxyRequest.Application = nil
	}

	_, err := basisTheoryClient.Proxies.Update(ctx, getStringValue(proxy.ID), updateProxyRequest)

	if err != nil {
		return apiErrorDiagnostics("Error updating Proxy:", err)
	}

	// Wait for the proxy to reach a final state before returning
	if diags := waitForProxyFinalState(ctx, basisTheoryClient, data.Id()); diags != nil {
		return diags
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
		ID:             &id,
		Name:           getStringPointer(data.Get("name")),
		DestinationURL: getStringPointer(data.Get("destination_url")),
		ApplicationID:  getStringPointer(data.Get("application_id")),
		RequireAuth:    getBoolPointer(data.Get("require_auth")),
	}

	// Handle request_transforms array
	proxy.RequestTransforms = parseTransformsFromData(data, "request_transforms")

	// Handle response_transforms array
	proxy.ResponseTransforms = parseTransformsFromData(data, "response_transforms")

	configOptions := map[string]*string{}
	if cfg, ok := data.GetOk("configuration"); ok {
		for key, value := range cfg.(map[string]interface{}) {
			configOptions[key] = getStringPointer(value)
		}
	}

	proxy.Configuration = configOptions

	return proxy
}

func parseTransformsFromData(data *schema.ResourceData, fieldName string) []*basistheory.ProxyTransform {
	transformsRaw, ok := data.GetOk(fieldName)
	if !ok {
		return nil
	}

	transformsList, ok := transformsRaw.([]interface{})
	if !ok {
		return nil
	}

	var transforms []*basistheory.ProxyTransform
	for _, item := range transformsList {
		if transformMap, ok := item.(map[string]interface{}); ok {
			transform := &basistheory.ProxyTransform{}

			// Basic fields
			if val, exists := transformMap["type"]; exists && val != nil {
				transform.Type = getStringPointer(val)
			}
			if val, exists := transformMap["code"]; exists && val != nil {
				transform.Code = getStringPointer(val)
			}
			if val, exists := transformMap["matcher"]; exists && !IsNilOrEmpty(val) {
				transform.Matcher = getStringPointer(val)
			}
			if val, exists := transformMap["expression"]; exists && val != nil {
				transform.Expression = getStringPointer(val)
			}
			if val, exists := transformMap["replacement"]; exists && val != nil {
				transform.Replacement = getStringPointer(val)
			}

			// Handle options (now as nested resource)
			if val, exists := transformMap["options"]; exists && val != nil {
				if optionsList, ok := val.([]interface{}); ok && len(optionsList) > 0 {
					if optionsMap, ok := optionsList[0].(map[string]interface{}); ok {
						options := &basistheory.ProxyTransformOptions{}

						// Existing tokenize/append fields
						if identifier, exists := optionsMap["identifier"]; exists && identifier != nil {
							options.Identifier = getStringPointer(identifier)
						}
						if value, exists := optionsMap["value"]; exists && value != nil {
							options.Value = getStringPointer(value)
						}
						if location, exists := optionsMap["location"]; exists && location != nil {
							options.Location = getStringPointer(location)
						}
						if token, exists := optionsMap["token"]; exists && token != nil {
							if tokenStr, ok := token.(string); ok && tokenStr != "" {
								var createTokenRequest basistheory.CreateTokenRequest
								if err := json.Unmarshal([]byte(tokenStr), &createTokenRequest); err == nil {
									options.Token = &createTokenRequest
								}
							}
						}

						// Handle runtime (now under options)
						if rtRaw, exists := optionsMap["runtime"]; exists && rtRaw != nil {
							if rtList, ok := rtRaw.([]interface{}); ok && len(rtList) > 0 {
								if rtMap, ok := rtList[0].(map[string]interface{}); ok {
									rt := &basistheory.Runtime{}
									if val, ok := rtMap["image"]; ok {
										rt.Image = getStringPointer(val)
									}
									if val, ok := rtMap["warm_concurrency"]; ok {
										rt.WarmConcurrency = getIntPointer(val)
									}
									if val, ok := rtMap["timeout"]; ok {
										rt.Timeout = getIntPointer(val)
									}
									if val, ok := rtMap["resources"]; ok {
										rt.Resources = getStringPointer(val)
									}
									if deps, ok := rtMap["dependencies"]; ok && deps != nil {
										depMap := map[string]*string{}
										for k, v := range deps.(map[string]interface{}) {
											depMap[k] = getStringPointer(v)
										}
										rt.Dependencies = depMap
									}
									if perms, ok := rtMap["permissions"]; ok && perms != nil {
										var ps []string
										for _, p := range perms.([]interface{}) {
											if s, ok := p.(string); ok {
												ps = append(ps, s)
											}
										}
										rt.Permissions = ps
									}
									options.Runtime = rt
								}
							}
						}

						// Only set Options if at least one field is provided
						if options.Identifier != nil || options.Value != nil || options.Location != nil || options.Token != nil || options.Runtime != nil {
							transform.Options = options
						}
					}
				}
			}

			transforms = append(transforms, transform)
		}
	}

	return transforms
}


func flattenProxyTransforms(transforms []*basistheory.ProxyTransform) []map[string]interface{} {
	if transforms == nil {
		return nil
	}

	result := make([]map[string]interface{}, 0, len(transforms))
	for _, transform := range transforms {
		if transform == nil {
			continue
		}

		flattenedTransform := map[string]interface{}{}

		// Basic fields
		if transform.Type != nil {
			flattenedTransform["type"] = *transform.Type
		}
		if transform.Code != nil {
			flattenedTransform["code"] = *transform.Code
		}
		if transform.Matcher != nil {
			flattenedTransform["matcher"] = *transform.Matcher
		}
		if transform.Expression != nil {
			flattenedTransform["expression"] = *transform.Expression
		}
		if transform.Replacement != nil {
			flattenedTransform["replacement"] = *transform.Replacement
		}

		// Handle options as nested resource
		if transform.Options != nil {
			options := transform.Options
			optionsMap := map[string]interface{}{}

			if options.Identifier != nil {
				optionsMap["identifier"] = *options.Identifier
			}
			if options.Value != nil {
				optionsMap["value"] = *options.Value
			}
			if options.Location != nil {
				optionsMap["location"] = *options.Location
			}
			if options.Token != nil {
				if tokenBytes, err := json.Marshal(options.Token); err == nil {
					optionsMap["token"] = string(tokenBytes)
				}
			}

			// Handle runtime under options - ONLY include if it has actual data
			if options.Runtime != nil {
				rt := options.Runtime
				rtMap := map[string]interface{}{}
				hasRuntimeData := false

				if rt.Image != nil && *rt.Image != "" {
					rtMap["image"] = *rt.Image
					hasRuntimeData = true
				}
				if rt.Dependencies != nil && len(rt.Dependencies) > 0 {
					deps := map[string]string{}
					for k, p := range rt.Dependencies {
						if p != nil {
							deps[k] = *p
						}
					}
					if len(deps) > 0 {
						rtMap["dependencies"] = deps
						hasRuntimeData = true
					}
				}
				if rt.WarmConcurrency != nil {
					rtMap["warm_concurrency"] = *rt.WarmConcurrency
					hasRuntimeData = true
				}
				if rt.Timeout != nil {
					rtMap["timeout"] = *rt.Timeout
					hasRuntimeData = true
				}
				if rt.Resources != nil && *rt.Resources != "" {
					rtMap["resources"] = *rt.Resources
					hasRuntimeData = true
				}
				if rt.Permissions != nil && len(rt.Permissions) > 0 {
					rtMap["permissions"] = rt.Permissions
					hasRuntimeData = true
				}

				// Only include runtime block if it has actual data
				if hasRuntimeData {
					optionsMap["runtime"] = []interface{}{rtMap}
				}
			}

			if len(optionsMap) > 0 {
				flattenedTransform["options"] = []interface{}{optionsMap}
			}
		}

		result = append(result, flattenedTransform)
	}

	return result
}

func validateRequestTransforms(val interface{}, _ string) (warns []string, errs []error) {
	transforms, ok := val.([]interface{})
	if !ok {
		errs = append(errs, fmt.Errorf("expected list of transforms"))
		return
	}

	return validateProxyTransforms(transforms, "request_transforms")
}

func validateResponseTransforms(val interface{}, _ string) (warns []string, errs []error) {
	transforms, ok := val.([]interface{})
	if !ok {
		errs = append(errs, fmt.Errorf("expected list of transforms"))
		return
	}

	return validateProxyTransforms(transforms, "response_transforms")
}

func validateProxyTransforms(transforms []interface{}, fieldName string) (warns []string, errs []error) {
	if len(transforms) == 0 {
		return
	}

	// Track identifiers for uniqueness validation
	identifiers := make(map[string]bool)
	codeTransformCount := 0

	for i, transformRaw := range transforms {
		transform, ok := transformRaw.(map[string]interface{})
		if !ok {
			errs = append(errs, fmt.Errorf("%s[%d]: expected transform object", fieldName, i))
			continue
		}

		// Validate individual transform
		transformWarns, transformErrs := validateProxyTransform(transform, fmt.Sprintf("%s[%d]", fieldName, i))
		warns = append(warns, transformWarns...)
		errs = append(errs, transformErrs...)

		// Check for code transforms
		if transformType, exists := transform["type"]; exists {
			if typeStr, ok := transformType.(string); ok && typeStr == "code" {
				codeTransformCount++
			}
		}

		// Check for duplicate identifiers in tokenize transforms
		if options, exists := transform["options"]; exists {
			// Handle both old format (map) and new format (list) during validation
			var optionsMap map[string]interface{}

			// Check if it's the old map format (for backward compatibility during migration)
			if oldOptionsMap, ok := options.(map[string]interface{}); ok {
				optionsMap = oldOptionsMap
			} else if optionsList, ok := options.([]interface{}); ok && len(optionsList) > 0 {
				// New format: list with single item
				if newOptionsMap, ok := optionsList[0].(map[string]interface{}); ok {
					optionsMap = newOptionsMap
				}
			}

			if optionsMap != nil {
				if identifier, exists := optionsMap["identifier"]; exists {
					if identifierStr, ok := identifier.(string); ok && identifierStr != "" {
						if identifiers[identifierStr] {
							errs = append(errs, fmt.Errorf("duplicate identifier found in %s: %s", fieldName, identifierStr))
						} else {
							identifiers[identifierStr] = true
						}
					}
				}
			}
		}
	}

	// Validate only one CODE transform is allowed
	if codeTransformCount > 1 {
		errs = append(errs, fmt.Errorf("only one CODE transform is allowed in %s", fieldName))
	}

	return
}

func validateProxyTransform(transform map[string]interface{}, fieldName string) (warns []string, errs []error) {
	// Basic type validation
	if transformType, exists := transform["type"]; exists {
		if typeStr, ok := transformType.(string); ok {
			switch typeStr {
			case "code":
				// Code validation
				if code, exists := transform["code"]; !exists || code == nil || code.(string) == "" {
					errs = append(errs, fmt.Errorf("%s: code is required when type is 'code'", fieldName))
				}
				// Code transforms should not have matcher, expression, or replacement
				if matcher, exists := transform["matcher"]; exists && matcher != nil && matcher.(string) != "" {
					errs = append(errs, fmt.Errorf("%s: matcher is not valid when type is 'code'", fieldName))
				}
				if expression, exists := transform["expression"]; exists && expression != nil && expression.(string) != "" {
					errs = append(errs, fmt.Errorf("%s: expression is not valid when type is 'code'", fieldName))
				}
				if replacement, exists := transform["replacement"]; exists && replacement != nil && replacement.(string) != "" {
					errs = append(errs, fmt.Errorf("%s: replacement is not valid when type is 'code'", fieldName))
				}
			case "mask":
				// Mask validation
				if matcher, exists := transform["matcher"]; !exists || matcher == nil || matcher.(string) == "" {
					errs = append(errs, fmt.Errorf("%s: matcher is required when type is 'mask'", fieldName))
				} else if matcherStr, ok := matcher.(string); ok {
					if matcherStr == "regex" {
						if expression, exists := transform["expression"]; !exists || expression == nil || expression.(string) == "" {
							errs = append(errs, fmt.Errorf("%s: expression is required when type is 'mask' and matcher is 'regex'", fieldName))
						}
					} else if matcherStr == "chase_stratus_pan" {
						if expression, exists := transform["expression"]; exists && expression != nil && expression.(string) != "" {
							errs = append(errs, fmt.Errorf("%s: expression should not be provided when matcher is 'chase_stratus_pan'", fieldName))
						}
					}
				}
				if replacement, exists := transform["replacement"]; !exists || replacement == nil || replacement.(string) == "" {
					errs = append(errs, fmt.Errorf("%s: replacement is required when type is 'mask'", fieldName))
				}
			case "tokenize":
				// Tokenize validation - handle both old and new format
				if options, exists := transform["options"]; !exists || options == nil {
					errs = append(errs, fmt.Errorf("%s: options are required for tokenize transforms", fieldName))
				} else {
					var optionsMap map[string]interface{}

					// Handle both old format (map) and new format (list)
					if oldOptionsMap, ok := options.(map[string]interface{}); ok {
						optionsMap = oldOptionsMap
					} else if optionsList, ok := options.([]interface{}); ok && len(optionsList) > 0 {
						if newOptionsMap, ok := optionsList[0].(map[string]interface{}); ok {
							optionsMap = newOptionsMap
						}
					}

					if optionsMap != nil {
						if token, exists := optionsMap["token"]; !exists || token == nil || token.(string) == "" {
							errs = append(errs, fmt.Errorf("%s: token is required in tokenize transform options", fieldName))
						} else if tokenStr, ok := token.(string); ok && tokenStr != "" {
							// Validate that token is valid JSON
							var js json.RawMessage
							if err := json.Unmarshal([]byte(tokenStr), &js); err != nil {
								errs = append(errs, fmt.Errorf("%s: token must be valid JSON: %s", fieldName, err))
							}
						}
						// Validate identifier format if provided
						if identifier, exists := optionsMap["identifier"]; exists && identifier != nil {
							if identifierStr, ok := identifier.(string); ok && identifierStr != "" {
								if len(identifierStr) > 100 {
									errs = append(errs, fmt.Errorf("%s: identifier must be 100 characters or less", fieldName))
								}
							}
						}
					}
				}
			case "append_json", "append_text", "append_header":
				// Append transform validation - handle both old and new format
				if options, exists := transform["options"]; !exists || options == nil {
					errs = append(errs, fmt.Errorf("%s: options are required for append transforms", fieldName))
				} else {
					var optionsMap map[string]interface{}

					// Handle both old format (map) and new format (list)
					if oldOptionsMap, ok := options.(map[string]interface{}); ok {
						optionsMap = oldOptionsMap
					} else if optionsList, ok := options.([]interface{}); ok && len(optionsList) > 0 {
						if newOptionsMap, ok := optionsList[0].(map[string]interface{}); ok {
							optionsMap = newOptionsMap
						}
					}

					if optionsMap != nil {
						if value, exists := optionsMap["value"]; !exists || value == nil || value.(string) == "" {
							errs = append(errs, fmt.Errorf("%s: value is required in append transform options", fieldName))
						}
						if typeStr == "append_json" || typeStr == "append_header" {
							if location, exists := optionsMap["location"]; !exists || location == nil || location.(string) == "" {
								errs = append(errs, fmt.Errorf("%s: location is required for %s transforms", fieldName, typeStr))
							}
						}
					}
				}
			}
		}
	}

	return
}

func IsNilOrEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	// Handle string pointers
	if strPtr, ok := value.(*string); ok {
		return strPtr == nil || *strPtr == ""
	}

	// Handle direct strings
	if str, ok := value.(string); ok {
		return str == ""
	}

	return false
}

// jsonEqual compares two JSON strings for semantic equality
func jsonEqual(a, b string) bool {
	if a == b {
		return true
	}

	// Parse both JSON strings
	var aObj, bObj interface{}
	if err := json.Unmarshal([]byte(a), &aObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &bObj); err != nil {
		return false
	}

	// Marshal both back to normalized JSON for comparison
	aBytes, err := json.Marshal(aObj)
	if err != nil {
		return false
	}
	bBytes, err := json.Marshal(bObj)
	if err != nil {
		return false
	}

	return string(aBytes) == string(bBytes)
}