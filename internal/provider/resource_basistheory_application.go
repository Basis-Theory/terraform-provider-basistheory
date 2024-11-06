package provider

import (
	"context"
	"regexp"

	"github.com/Basis-Theory/basistheory-go/v6"
	basistheoryV2 "github.com/Basis-Theory/go-sdk"
	basistheoryV2client "github.com/Basis-Theory/go-sdk/client"
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
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	application := getApplicationFromDataV2(data)

	createApplicationRequest := &basistheoryV2.CreateApplicationRequest{
		Name: getStringValue(application.Name),
		Type: getStringValue(application.Type),
		Permissions: application.Permissions,
		Rules: application.Rules,
		CreateKey: getBoolPointer(data.Get("create_key")),
	}

	createdApplication, err := basisTheoryClient.Applications.Create(ctx, createApplicationRequest)

	if err != nil {
		return apiErrorDiagnosticsV2("Error creating Application:", err)
	}

	data.SetId(*createdApplication.ID)
	createdApplicationKeys := createdApplication.Keys

	if len(createdApplicationKeys) > 0 {
		err = data.Set("key", createdApplicationKeys[0].Key)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceApplicationRead(ctx, data, meta)
}

func resourceApplicationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	application, err := basisTheoryClient.Applications.Get(ctx, data.Id())

	if err != nil {
		return apiErrorDiagnosticsV2("Error reading Application:", err)
	}

	data.SetId(*application.ID)

	permissions := application.Permissions
	rules := application.Rules

	modifiedAt := ""

	if application.ModifiedAt != nil {
		modifiedAt = application.ModifiedAt.String()
	}

	for applicationDatumName, applicationDatum := range map[string]interface{}{
		"tenant_id":   application.TenantID,
		"name":        application.Name,
		"type":        application.Type,
		"permissions": permissions,
		"rule":        flattenAccessRuleData(rules),
		"created_at":  application.CreatedAt.String(),
		"created_by":  application.CreatedBy,
		"modified_at": modifiedAt,
		"modified_by": application.ModifiedBy,
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

func getApplicationFromDataV2(data *schema.ResourceData) basistheoryV2.Application {
	var permissions []string
	if dataPermissions, ok := data.Get("permissions").(*schema.Set); ok {
		for _, dataPermission := range dataPermissions.List() {
			permissions = append(permissions, dataPermission.(string))
		}
	}

	var rules []*basistheoryV2.AccessRule
	if dataRules, ok := data.Get("rule").(*schema.Set); ok {
		for _, dataRule := range dataRules.List() {
			ruleMap := dataRule.(map[string]interface{})

			var rulePermissions []string
			if dataRulePermissions, ok := ruleMap["permissions"].(*schema.Set); ok {
				for _, dataRulePermission := range dataRulePermissions.List() {
					rulePermissions = append(rulePermissions, dataRulePermission.(string))
				}
			}
			rule := &basistheoryV2.AccessRule {
				Description: getStringPointer(ruleMap["description"]),
				Priority: getIntPointer(ruleMap["priority"]),
				Container: getStringPointer(ruleMap["container"]),
				Transform: getStringPointer(ruleMap["transform"]),
				Permissions: rulePermissions,
			}

			rules = append(rules, rule)
		}
	}

	id := data.Id()
	application := basistheoryV2.Application {
		ID: &id,
		Name: getStringPointer(data.Get("name")),
		TenantID: getStringPointer(data.Get("tenant_id")),
		Type: getStringPointer(data.Get("type")),
		Permissions: permissions,
		Rules: rules,
	}

	return application
}

func flattenAccessRuleData(accessRules []*basistheoryV2.AccessRule) []interface{} {
	if accessRules != nil {
		var flattenedAccessRules []interface{}

		for _, rule := range accessRules {
			flattenedAccessRule := make(map[string]interface{})

			flattenedAccessRule["description"] = rule.Description
			flattenedAccessRule["priority"] = rule.Priority
			flattenedAccessRule["container"] = rule.Container
			flattenedAccessRule["transform"] = rule.Transform
			flattenedAccessRule["permissions"] = rule.Permissions

			flattenedAccessRules = append(flattenedAccessRules, flattenedAccessRule)
		}

		return flattenedAccessRules
	}

	return make([]interface{}, 0)
}

func getStringPointer(value interface{}) *string {
	if str, ok := value.(string); ok {
		return &str
	}
	return nil
}

func getStringValue(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}

func getBoolPointer(value interface{}) *bool {
	if b, ok := value.(bool); ok {
		return &b
	}
	return nil
}

func getIntPointer(value interface{}) *int {
	if i, ok := value.(int); ok {
		return &i
	}
	return nil
}