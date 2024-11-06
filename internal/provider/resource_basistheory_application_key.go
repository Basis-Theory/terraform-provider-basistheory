package provider

import (
	"context"
	basistheoryV2client "github.com/Basis-Theory/go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
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
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	applicationId := data.Get("application_id").(string)

	createdApplicationKey, err := basisTheoryClient.ApplicationKeys.Create(ctx, applicationId)

	if err != nil {
		return apiErrorDiagnosticsV2("Error creating ApplicationKey:", err)
	}

	data.SetId(*createdApplicationKey.ID)

	for datumName, datumValue := range map[string]interface{}{
		"application_id": applicationId,
		"key":            createdApplicationKey.Key,
		"created_at":     createdApplicationKey.CreatedAt.String(),
		"created_by":     createdApplicationKey.CreatedBy,
	} {
		err := data.Set(datumName, datumValue)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceApplicationKeyRead(ctx, data, meta)
}

func resourceApplicationKeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	applicationId := data.Get("application_id").(string)
	applicationKey, err := basisTheoryClient.ApplicationKeys.Get(ctx, applicationId, data.Id())

	if err != nil {
		return apiErrorDiagnosticsV2("Error reading ApplicationKey:", err)
	}

	data.SetId(*applicationKey.ID)

	for datumName, datumValue := range map[string]interface{}{
		"created_at": applicationKey.CreatedAt.String(),
		"created_by": applicationKey.CreatedBy,
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
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	applicationId := data.Get("application_id").(string)
	keyId := data.Id()
	err := basisTheoryClient.ApplicationKeys.Delete(ctx, applicationId, keyId)

	if err != nil && !strings.Contains(err.Error(), "Not Found") {
		return apiErrorDiagnosticsV2("Error deleting ApplicationKey appId: ", err)
	}

	return nil
}
