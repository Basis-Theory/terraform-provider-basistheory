package provider

import (
	"context"
	"github.com/Basis-Theory/basistheory-go/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryApplicationKey() *schema.Resource {
	return &schema.Resource{
		Description: "Application Keys https://developers.basistheory.com/docs/api/applications/keys",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceApplicationKeyCreate,
		ReadContext:   resourceApplicationKeyRead,
		UpdateContext: resourceApplicationKeyUpdate,
		DeleteContext: resourceApplicationKeyDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Application Key",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"application_id": {
				Description: "Application identifier where this Application Key was created",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key": {
				Description: "Key for the Application Key",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
			"created_at": {
				Description: "Timestamp at which the Application Key was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Application Key",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceApplicationKeyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	applicationId := data.Get("application_id").(string)

	createdApplicationKey, response, err := basisTheoryClient.ApplicationKeysApi.Create(ctxWithApiKey, applicationId).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error creating ApplicationKey:", response, err)
	}

	data.SetId(createdApplicationKey.GetId())

	for datumName, datumValue := range map[string]interface{}{
		"application_id": applicationId,
		"key":            createdApplicationKey.GetKey(),
		"created_at":     createdApplicationKey.GetCreatedAt().String(),
		"created_by":     createdApplicationKey.GetCreatedBy(),
	} {
		err := data.Set(datumName, datumValue)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceApplicationKeyRead(ctx, data, meta)
}

func resourceApplicationKeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	applicationId := data.Get("application_id").(string)
	applicationKey, response, err := basisTheoryClient.ApplicationKeysApi.GetById(ctxWithApiKey, applicationId, data.Id()).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error reading ApplicationKey:", response, err)
	}

	data.SetId(applicationKey.GetId())

	for datumName, datumValue := range map[string]interface{}{
		"created_at": applicationKey.GetCreatedAt().String(),
		"created_by": applicationKey.GetCreatedBy(),
	} {
		err := data.Set(datumName, datumValue)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceApplicationKeyUpdate(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	oldAppId, _ := data.GetChange("application_id")

	err := data.Set("application_id", oldAppId)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.Errorf("Updating ApplicationKey is not supported.")
}

func resourceApplicationKeyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ctxWithApiKey := getContextWithApiKey(ctx, meta.(map[string]interface{})["api_key"].(string))
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheory.APIClient)

	applicationId := data.Get("application_id").(string)
	keyId := data.Id()
	response, err := basisTheoryClient.ApplicationKeysApi.Delete(ctxWithApiKey, applicationId, keyId).Execute()

	if err != nil {
		return apiErrorDiagnostics("Error deleting ApplicationKey appId: ", response, err)
	}

	return nil
}
