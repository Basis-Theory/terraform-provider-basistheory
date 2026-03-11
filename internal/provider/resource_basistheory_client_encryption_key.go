package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	basistheory "github.com/Basis-Theory/go-sdk/v5"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryClientEncryptionKey() *schema.Resource {
	return &schema.Resource{
		Description: "Client Encryption Key https://developers.basistheory.com/docs/api/client-keys",
		
		DeprecationMessage: "The client_encryption_key resource is deprecated and will be removed in the next major version of the provider. " +
        			"Client encryption keys expire after 6 months by default, which causes state drift issues in Terraform. " +
        			"Please manage these keys outside of Terraform using the Basis Theory API or SDK.",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceClientEncryptionKeyCreate,
		ReadContext:   resourceClientEncryptionKeyRead,
		DeleteContext: resourceClientEncryptionKeyDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Client Encryption Key",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"expires_at": {
				Description: "Timestamp at which the Client Encryption Key will expire",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}

func resourceClientEncryptionKeyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	req := &basistheory.ClientEncryptionKeyRequest{}
	if v, ok := data.GetOk("expires_at"); ok {
		if s, ok := v.(string); ok && s != "" {
			parsedTime, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return apiErrorDiagnostics("Error parsing expires_at:", err)
			}
			req.ExpiresAt = &parsedTime
		}
	}

	key, err := basisTheoryClient.Keys.Create(ctx, req)
	if err != nil {
		return apiErrorDiagnostics("Error creating Client Encryption Key:", err)
	}

	if key.KeyID != nil {
		data.SetId(*key.KeyID)
	}
	return resourceClientEncryptionKeyRead(ctx, data, meta)
}

func resourceClientEncryptionKeyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	key, err := basisTheoryClient.Keys.Get(ctx, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			data.SetId("")
			return diag.Diagnostics{
				{
					Severity: diag.Warning,
					Summary:  "Client Encryption Key removed from state",
					Detail:   "The Client Encryption Key was not found (it may have expired or been deleted outside of Terraform) and has been removed from state. It will be recreated on the next apply.",
				},
			}
		}

		return apiErrorDiagnostics("Error reading Client Encryption Key:", err)
	}

	if key.KeyID != nil {
		data.SetId(*key.KeyID)
	}
	if key.ExpiresAt != nil {
		data.Set("expires_at", key.ExpiresAt.Format(time.RFC3339))
	}

	var diags diag.Diagnostics
	if key.ExpiresAt != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Client Encryption Key expiration",
			Detail:   fmt.Sprintf("This Client Encryption Key will be automatically deleted at %s. After expiration, it will be removed from Terraform state on the next plan/apply.", key.ExpiresAt.Format(time.RFC3339)),
		})
	} else {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Client Encryption Key expiration",
			Detail:   "No expiration date is set. This Client Encryption Key will be automatically deleted after 6 months by default.",
		})
	}
	return diags
}

func resourceClientEncryptionKeyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	err := basisTheoryClient.Keys.Delete(ctx, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}
		return apiErrorDiagnostics("Error deleting Client Encryption Key:", err)
	}

	return nil
}
