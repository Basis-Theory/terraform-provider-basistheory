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
							Description: "Options for tokenize and append transforms",
							Type:        schema.TypeMap,
							Optional:    true,
							Elem: &schema.Schema{Type: schema.TypeString},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// For token field, suppress diff if JSON content is equivalent
								if strings.HasSuffix(k, ".token") {
									return jsonEqual(old, new)
								}
								return old == new
							},
						},
						"runtime": {
							Description: "Runtime configuration for code transforms",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
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
							Description: "Options for tokenize and append transforms",
							Type:        schema.TypeMap,
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// For token field, suppress diff if JSON content is equivalent
								if strings.HasSuffix(k, ".token") {
									return jsonEqual(old, new)
								}
								return old == new
							},
						},
						"runtime": {
							Description: "Runtime configuration for code transforms",
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
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
			},
		},
	}
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

   // Options as map (legacy tokenize/append support)
			if val, exists := transformMap["options"]; exists && val != nil {
				if optionsMap, ok := val.(map[string]interface{}); ok {
					options := &basistheory.ProxyTransformOptions{}
					if identifier, exists := optionsMap["identifier"]; exists && identifier != nil { options.Identifier = getStringPointer(identifier) }
					if value, exists := optionsMap["value"]; exists && value != nil { options.Value = getStringPointer(value) }
					if location, exists := optionsMap["location"]; exists && location != nil { options.Location = getStringPointer(location) }
					if token, exists := optionsMap["token"]; exists && token != nil {
						if tokenStr, ok := token.(string); ok && tokenStr != "" {
							var createTokenRequest basistheory.CreateTokenRequest
							if err := json.Unmarshal([]byte(tokenStr), &createTokenRequest); err == nil {
								options.Token = &createTokenRequest
							}
						}
					}
					// Only set Options if at least one field is provided
					if options.Identifier != nil || options.Value != nil || options.Location != nil || options.Token != nil {
						transform.Options = options
					}
				}
			}

			// Transform-level runtime block (primary path used by tests)
			if rtRaw, ok := transformMap["runtime"]; ok && rtRaw != nil {
				if list, ok := rtRaw.([]interface{}); ok && len(list) > 0 {
					if m, ok := list[0].(map[string]interface{}); ok {
						rt := &basistheory.Runtime{}
						if val, ok := m["image"]; ok { rt.Image = getStringPointer(val) }
						if val, ok := m["warm_concurrency"]; ok { rt.WarmConcurrency = getIntPointer(val) }
						if val, ok := m["timeout"]; ok { rt.Timeout = getIntPointer(val) }
						if val, ok := m["resources"]; ok { rt.Resources = getStringPointer(val) }
						if deps, ok := m["dependencies"]; ok && deps != nil {
							depMap := map[string]*string{}
							for k, v := range deps.(map[string]interface{}) { depMap[k] = getStringPointer(v) }
							rt.Dependencies = depMap
						}
						if perms, ok := m["permissions"]; ok && perms != nil {
							var ps []string
							for _, p := range perms.([]interface{}) { if s, ok := p.(string); ok { ps = append(ps, s) } }
							rt.Permissions = ps
						}
      // Assign runtime at the transform level
						if transform.Options == nil { transform.Options = &basistheory.ProxyTransformOptions{} }
						transform.Options.Runtime = rt
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

  // Handle options (legacy map-style fields)
	if transform.Options != nil {
		options := map[string]interface{}{}
		if transform.Options.Identifier != nil { options["identifier"] = *transform.Options.Identifier }
		if transform.Options.Value != nil { options["value"] = *transform.Options.Value }
		if transform.Options.Location != nil { options["location"] = *transform.Options.Location }
		if transform.Options.Token != nil {
			if tokenBytes, err := json.Marshal(transform.Options.Token); err == nil {
				options["token"] = string(tokenBytes)
			}
		}
		if len(options) > 0 {
			flattenedTransform["options"] = options
		}
	}
	// Expose runtime as a nested block at transform level to follow reactor example
	var rtSource *basistheory.Runtime
	if transform != nil {
		if transform.Options != nil && transform.Options.Runtime != nil {
			rtSource = transform.Options.Runtime
		}
	}
	if rtSource != nil {
		rt := rtSource
		rtMap := map[string]interface{}{}
		if rt.Image != nil { rtMap["image"] = *rt.Image }
		if rt.Dependencies != nil {
			deps := map[string]string{}
			for k, p := range rt.Dependencies { if p != nil { deps[k] = *p } }
			rtMap["dependencies"] = deps
		}
		if rt.WarmConcurrency != nil { rtMap["warm_concurrency"] = *rt.WarmConcurrency }
		if rt.Timeout != nil { rtMap["timeout"] = *rt.Timeout }
		if rt.Resources != nil { rtMap["resources"] = *rt.Resources }
		if rt.Permissions != nil { rtMap["permissions"] = rt.Permissions }
		flattenedTransform["runtime"] = []interface{}{rtMap}
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
			if optionsMap, ok := options.(map[string]interface{}); ok {
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
				// Tokenize validation
				if options, exists := transform["options"]; !exists || options == nil {
					errs = append(errs, fmt.Errorf("%s: options are required for tokenize transforms", fieldName))
				} else if optionsMap, ok := options.(map[string]interface{}); ok {
					if token, exists := optionsMap["token"]; !exists || token == nil {
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
			case "append_json", "append_text", "append_header":
				// Append transform validation
				if options, exists := transform["options"]; !exists || options == nil {
					errs = append(errs, fmt.Errorf("%s: options are required for append transforms", fieldName))
				} else if optionsMap, ok := options.(map[string]interface{}); ok {
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
