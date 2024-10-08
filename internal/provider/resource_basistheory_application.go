package provider

import (
	"context"
	"regexp"

	"github.com/Basis-Theory/basistheory-go/v6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBasisTheoryApplication() *schema.Resource {
	var (
		applicationTypes     = []string{"public", "private", "management"}
		accessRuleTransforms = []string{"mask", "redact", "reveal"}
	)

	return &schema.Resource{
		Description: "Application https://developers.basistheory.com/docs/api/applications",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceApplicationCreate,
		ReadContext:   resourceApplicationRead,
		UpdateContext: resourceApplicationUpdate,
		DeleteContext: resourceApplicationDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Application",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the Application",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key": {
				Description: "Key for the Application",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"tenant_id": {
				Description: "Tenant identifier where this Application was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"type": {
				Description:  "Type for the Application",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(applicationTypes, false),
			},
			"create_key": {
				Description: "Create Application Key by default. We suggest omitting 'create_key' and manage API Keys with the 'basistheory_application_key' resource",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"permissions": {
				Description: "Permissions for the Application",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"rule": {
				Description: "Access rules for the Application",
				Type:        schema.TypeSet,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Description:  "A description of this Access Rule",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringMatch(regexp.MustCompile("^[A-Z-_]+"), "Configuration name can only contain uppercase letters, '-', and '_'"),
						},
						"priority": {
							Description:  "Description of what the configuration option is for and/or possible values",
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"container": {
							Description: "The container of Tokens this rule is scoped to",
							Type:        schema.TypeString,
							Required:    true,
						},
						"transform": {
							Description:  "The transform to apply to accessed Tokens",
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(accessRuleTransforms, false),
						},
						"permissions": {
							Description: "List of permissions to grant on this Access Rule",
							Type:        schema.TypeSet,
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"created_at": {
				Description: "Timestamp at which the Application was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Application",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_at": {
				Description: "Timestamp at which the Application was last updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_by": {
				Description: "Identifier for who last modified the Application",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    applicationInstanceResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: applicationInstanceStateUpgradeV0,
				Version: 0,
			},
		},
	}
}

func applicationInstanceResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"create_key": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func applicationInstanceStateUpgradeV0(_ context.Context, rawState map[string]any, _ any) (map[string]any, error) {
	rawState["create_key"] = "true"

	return rawState, nil
}

func resourceApplicationCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	application := getApplicationFromData(data)

	createApplicationRequest := *basistheory.NewCreateApplicationRequest(application.GetName(), application.GetType())
	createApplicationRequest.SetPermissions(application.GetPermissions())
	createApplicationRequest.SetRules(application.GetRules())
	createApplicationRequest.SetCreateKey(data.Get("create_key").(bool))

	createdApplication, response, err := basisTheoryClient.ApplicationsApi.Create(ctxWithApiKey).CreateApplicationRequest(createApplicationRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Application:", response, err)
	}

	data.SetId(createdApplication.GetId())
	createdApplicationKeys := createdApplication.GetKeys()

	if len(createdApplicationKeys) > 0 {
		err = data.Set("key", createdApplicationKeys[0].GetKey())
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceApplicationRead(ctx, data, meta)
}

func resourceApplicationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	application, response, err := basisTheoryClient.ApplicationsApi.GetById(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading Application:", response, err)
	}

	data.SetId(application.GetId())

	permissions := application.GetPermissions()
	rules := application.GetRules()

	modifiedAt := ""

	if application.ModifiedAt.IsSet() {
		modifiedAt = application.GetModifiedAt().String()
	}

	for applicationDatumName, applicationDatum := range map[string]interface{}{
		"tenant_id":   application.GetTenantId(),
		"name":        application.GetName(),
		"type":        application.GetType(),
		"permissions": permissions,
		"rule":        flattenAccessRuleData(rules),
		"created_at":  application.GetCreatedAt().String(),
		"created_by":  application.GetCreatedBy(),
		"modified_at": modifiedAt,
		"modified_by": application.GetModifiedBy(),
	} {
		err := data.Set(applicationDatumName, applicationDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceApplicationUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if data.HasChange("create_key") {
		oldCreateKey, _ := data.GetChange("create_key")
		err := data.Set("create_key", oldCreateKey)

		if err != nil {
			return diag.FromErr(err)
		}

		return diag.Errorf("Updating 'create_key' is not supported.")
	}

	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	application := getApplicationFromData(data)

	updateApplicationRequest := *basistheory.NewUpdateApplicationRequest(application.GetName())
	updateApplicationRequest.SetPermissions(application.GetPermissions())
	updateApplicationRequest.SetRules(application.GetRules())

	_, response, err := basisTheoryClient.ApplicationsApi.Update(ctxWithApiKey, application.GetId()).UpdateApplicationRequest(updateApplicationRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Application:", response, err)
	}

	return resourceApplicationRead(ctx, data, meta)
}

func resourceApplicationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	response, err := basisTheoryClient.ApplicationsApi.Delete(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error deleting Application:", response, err)
	}

	return nil
}

func getApplicationFromData(data *schema.ResourceData) *basistheory.Application {
	application := basistheory.NewApplication()
	application.SetId(data.Id())
	application.SetName(data.Get("name").(string))
	application.SetTenantId(data.Get("tenant_id").(string))
	application.SetType(data.Get("type").(string))

	var permissions []string
	if dataPermissions, ok := data.Get("permissions").(*schema.Set); ok {
		for _, dataPermission := range dataPermissions.List() {
			permissions = append(permissions, dataPermission.(string))
		}
	}

	application.SetPermissions(permissions)

	var rules []basistheory.AccessRule
	if dataRules, ok := data.Get("rule").(*schema.Set); ok {
		for _, dataRule := range dataRules.List() {
			ruleMap := dataRule.(map[string]interface{})
			rule := *basistheory.NewAccessRule()
			rule.SetDescription(ruleMap["description"].(string))
			rule.SetPriority(int32(ruleMap["priority"].(int)))
			rule.SetContainer(ruleMap["container"].(string))
			rule.SetTransform(ruleMap["transform"].(string))

			var rulePermissions []string
			if dataRulePermissions, ok := ruleMap["permissions"].(*schema.Set); ok {
				for _, dataRulePermission := range dataRulePermissions.List() {
					rulePermissions = append(rulePermissions, dataRulePermission.(string))
				}
			}
			rule.SetPermissions(rulePermissions)
			rules = append(rules, rule)
		}
	}

	application.SetRules(rules)

	return application
}

func flattenAccessRuleData(accessRules []basistheory.AccessRule) []interface{} {
	if accessRules != nil {
		var flattenedAccessRules []interface{}

		for _, rule := range accessRules {
			flattenedAccessRule := make(map[string]interface{})

			flattenedAccessRule["description"] = rule.GetDescription()
			flattenedAccessRule["priority"] = rule.GetPriority()
			flattenedAccessRule["container"] = rule.GetContainer()
			flattenedAccessRule["transform"] = rule.GetTransform()
			flattenedAccessRule["permissions"] = rule.GetPermissions()

			flattenedAccessRules = append(flattenedAccessRules, flattenedAccessRule)
		}

		return flattenedAccessRules
	}

	return make([]interface{}, 0)
}
