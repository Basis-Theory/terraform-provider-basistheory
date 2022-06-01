package provider

import (
	"context"
	"github.com/Basis-Theory/basistheory-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceBasisTheoryApplication() *schema.Resource {
	var applicationTypes = []string{"elements", "public", "server_to_server", "management"}

	return &schema.Resource{
		Description: "Application https://docs.basistheory.com/#applications",

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
			"permissions": {
				Description: "Permissions for the Application",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
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
	}
}

func resourceApplicationCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	application := getApplicationFromData(data)

	createApplicationRequest := *basistheory.NewCreateApplicationRequest(application.GetName(), application.GetType())
	createApplicationRequest.SetPermissions(application.GetPermissions())

	createdApplication, response, err := basisTheoryClient.ApplicationsApi.ApplicationsCreate(ctxWithApiKey).CreateApplicationRequest(createApplicationRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating Application:", response, err)
	}

	data.SetId(createdApplication.GetId())
	err = data.Set("key", createdApplication.GetKey())

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceApplicationRead(ctx, data, meta)
}

func resourceApplicationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	application, response, err := basisTheoryClient.ApplicationsApi.ApplicationsGetById(ctxWithApiKey, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading Application:", response, err)
	}

	data.SetId(application.GetId())

	permissions := application.GetPermissions()

	modifiedAt := ""

	if application.ModifiedAt.IsSet() {
		modifiedAt = application.GetModifiedAt().String()
	}

	for applicationDatumName, applicationDatum := range map[string]interface{}{
		"tenant_id":   application.GetTenantId(),
		"name":        application.GetName(),
		"type":        application.GetType(),
		"permissions": permissions,
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
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	application := getApplicationFromData(data)
	updateApplicationRequest := *basistheory.NewUpdateApplicationRequest(application.GetName())
	updateApplicationRequest.SetPermissions(application.GetPermissions())

	_, response, err := basisTheoryClient.ApplicationsApi.ApplicationsUpdate(ctxWithApiKey, application.GetId()).UpdateApplicationRequest(updateApplicationRequest).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error updating Application:", response, err)
	}

	return resourceApplicationRead(ctx, data, meta)
}

func resourceApplicationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	response, err := basisTheoryClient.ApplicationsApi.ApplicationsDelete(ctxWithApiKey, data.Id()).Execute()

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
	if dataConfig, ok := data.Get("permissions").(*schema.Set); ok {
		for _, permission := range dataConfig.List() {
			permissions = append(permissions, permission.(string))
		}
	}

	application.SetPermissions(permissions)

	return application
}
