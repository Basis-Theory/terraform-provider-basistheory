package provider

import (
	"context"
	basistheory "github.com/Basis-Theory/go-sdk/v4"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v4/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func resourceBasisTheoryReactor() *schema.Resource {
	return &schema.Resource{
		Description: "Reactor https://docs.basistheory.com/docs/api/reactors",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceReactorCreate,
		ReadContext:   resourceReactorRead,
		UpdateContext: resourceReactorUpdate,
		DeleteContext: resourceReactorDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Reactor",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the Reactor",
				Type:        schema.TypeString,
				Required:    true,
			},
			"code": {
				Description: "The code that is executed when the Reactor runs",
				Type:        schema.TypeString,
				Required:    true,
			},
			"application_id": {
				Description: "The Application's permissions used in the BasisTheory instance passed into the Reactor",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},
			"tenant_id": {
				Description: "Tenant identifier where this Reactor was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"configuration": {
				Description: "Configuration for the Reactor",
				Type:        schema.TypeMap,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"runtime": {
				Description: "Runtime configuration for the Reactor",
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image": {
							Description: "Runtime image (e.g., node22)",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"dependencies": {
							Description: "Runtime dependencies",
							Type:        schema.TypeMap,
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"warm_concurrency": {
							Description: "Warm concurrency setting",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"timeout": {
							Description: "Timeout setting in seconds",
							Type:        schema.TypeInt,
							Optional:    true,
						},
						"resources": {
							Description: "Resource allocation (e.g., standard)",
							Type:        schema.TypeString,
							Optional:    true,
						},
						"permissions": {
							Description: "List of permissions for the reactor",
							Type:        schema.TypeList,
							Optional:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"state": {
				Description: "Current state of the Reactor",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Timestamp at which the Reactor was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Reactor",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_at": {
				Description: "Timestamp at which the Reactor was last updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_by": {
				Description: "Identifier for who last modified the Reactor",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceReactorCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	reactor := getReactorFromData(data)

 createReactorRequest := &basistheory.CreateReactorRequest{
		Name:          getStringValue(reactor.Name),
		Code:          getStringValue(reactor.Code),
		Configuration: reactor.Configuration,
		Application:   reactor.Application,
		Runtime:       reactor.Runtime,
	}

	createdReactor, err := basisTheoryClient.Reactors.Create(ctx, createReactorRequest)

	if err != nil {
		return apiErrorDiagnostics("Error creating Reactor:", err)
	}

	data.SetId(*createdReactor.ID)

	// Wait for the reactor to reach a final state before returning
	if diags := waitForReactorFinalState(ctx, basisTheoryClient, data.Id()); diags != nil {
		return diags
	}

	return resourceReactorRead(ctx, data, meta)
}

func waitForReactorFinalState(ctx context.Context, client *basistheoryClient.Client, id string) diag.Diagnostics {
	// Poll every 2 seconds up to 10 minutes
	interval := 2 * time.Second
	deadline := time.Now().Add(10 * time.Minute)

	for {
		if time.Now().After(deadline) {
			return diag.Errorf("timeout waiting for reactor %s to reach a final state", id)
		}

		reactor, err := client.Reactors.Get(ctx, id)
		if err != nil {
			return apiErrorDiagnostics("Error polling Reactor:", err)
		}

		state := ""
		if reactor.State != nil {
			state = *reactor.State
		}

		switch state {
		case "active", "outdated":
			return nil
		case "failed":
			return diag.Errorf("reactor %s reached failed state", id)
		}

		select {
		case <-ctx.Done():
			return diag.FromErr(ctx.Err())
		case <-time.After(interval):
		}
	}
}

func resourceReactorRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	reactor, err := basisTheoryClient.Reactors.Get(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnostics("Error reading Reactor:", err)
	}

	data.SetId(*reactor.ID)

	application := reactor.Application

	modifiedAt := ""

	if reactor.ModifiedAt != nil {
		modifiedAt = reactor.ModifiedAt.String()
	}

	runtime := reactor.Runtime

	// Flatten runtime to Terraform state (single block)
	runtimeMap := map[string]interface{}{}
	if runtime != nil {
		if v := runtime.Image; v != nil {
			runtimeMap["image"] = *v
		}
		if v := runtime.Dependencies; v != nil {
			deps := map[string]string{}
			for k, p := range v {
				if p != nil {
					deps[k] = *p
				}
			}
			runtimeMap["dependencies"] = deps
		}
		if v := runtime.WarmConcurrency; v != nil {
			runtimeMap["warm_concurrency"] = *v
		}
		if v := runtime.Timeout; v != nil {
			runtimeMap["timeout"] = *v
		}
		if v := runtime.Resources; v != nil {
			runtimeMap["resources"] = *v
		}
		if v := runtime.Permissions; v != nil {
			runtimeMap["permissions"] = v
		}
	}
	if len(runtimeMap) > 0 {
		if err := data.Set("runtime", []interface{}{runtimeMap}); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Clear runtime if not present
		if err := data.Set("runtime", nil); err != nil {
			return diag.FromErr(err)
		}
	}

	for reactorDatumName, reactorDatum := range map[string]interface{}{
		"tenant_id": reactor.TenantID,
		"name":      reactor.Name,
		"code":      reactor.Code,
		"application_id": func() interface{} {
			if application != nil {
				return application.ID
			}
			return nil
		}(),
		"configuration": reactor.Configuration,
		"state":         reactor.State,
		"created_at":    reactor.CreatedAt.String(),
		"created_by":    reactor.CreatedBy,
		"modified_at":   modifiedAt,
		"modified_by":   reactor.ModifiedBy,
	} {
		err := data.Set(reactorDatumName, reactorDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceReactorUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	reactor := getReactorFromData(data)
 updateReactorRequest := &basistheory.UpdateReactorRequest{
		Name:          getStringValue(reactor.Name),
		Code:          getStringValue(reactor.Code),
		Configuration: reactor.Configuration,
		Application:   reactor.Application,
		Runtime:       reactor.Runtime,
	}
	_, err := basisTheoryClient.Reactors.Update(ctx, *reactor.ID, updateReactorRequest)

	if err != nil {
		return apiErrorDiagnostics("Error updating Reactor:", err)
	}

	// Wait for the reactor to reach a final state before returning
	if diags := waitForReactorFinalState(ctx, basisTheoryClient, data.Id()); diags != nil {
		return diags
	}

	return resourceReactorRead(ctx, data, meta)
}

func resourceReactorDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	err := basisTheoryClient.Reactors.Delete(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnostics("Error deleting Reactor:", err)
	}

	return nil
}

func getReactorFromData(data *schema.ResourceData) *basistheory.Reactor {
	reactor := &basistheory.Reactor{}
	reactor.ID = getStringPointer(data.Id())
	reactor.Name = getStringPointer(data.Get("name"))

	reactorCode := data.Get("code").(string)
	if reactorCode != "" {
		reactor.Code = getStringPointer(reactorCode)
	}

	configOptions := map[string]*string{}
	if cfg, ok := data.GetOk("configuration"); ok {
		for key, value := range cfg.(map[string]interface{}) {
			configOptions[key] = getStringPointer(value)
		}
	}
	reactor.Configuration = configOptions

	// Application
	applicationId := data.Get("application_id").(string)
	if applicationId != "" {
		application := &basistheory.Application{}
		application.ID = getStringPointer(applicationId)
		reactor.Application = application
	}

	// Runtime block (max 1)
	if v, ok := data.GetOk("runtime"); ok {
		if list, ok := v.([]interface{}); ok && len(list) > 0 {
			if m, ok := list[0].(map[string]interface{}); ok {
				rt := &basistheory.Runtime{}
				if val, ok := m["image"]; ok {
					rt.Image = getStringPointer(val)
				}
				if val, ok := m["warm_concurrency"]; ok {
					rt.WarmConcurrency = getIntPointer(val)
				}
				if val, ok := m["timeout"]; ok {
					rt.Timeout = getIntPointer(val)
				}
				if val, ok := m["resources"]; ok {
					rt.Resources = getStringPointer(val)
				}
				// dependencies map[string]string -> map[string]*string
				if deps, ok := m["dependencies"]; ok && deps != nil {
					depMap := map[string]*string{}
					for k, v := range deps.(map[string]interface{}) {
						depMap[k] = getStringPointer(v)
					}
					rt.Dependencies = depMap
				}
				// permissions []interface{} -> []string
				if perms, ok := m["permissions"]; ok && perms != nil {
					var ps []string
					for _, p := range perms.([]interface{}) {
						if s, ok := p.(string); ok {
							ps = append(ps, s)
						}
					}
					rt.Permissions = ps
				}
				reactor.Runtime = rt
			}
		}
	}

	return reactor
}
