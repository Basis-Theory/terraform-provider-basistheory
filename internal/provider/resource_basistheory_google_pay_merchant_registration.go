package provider

import (
	"context"
	"errors"

	basistheory "github.com/Basis-Theory/go-sdk/v5"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/Basis-Theory/go-sdk/v5/googlepay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryGooglePayMerchantRegistration() *schema.Resource {
	return &schema.Resource{
		Description: "Google Pay Merchant Registration https://developers.basistheory.com/docs/api/google-pay/merchant-registration",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceGooglePayMerchantRegistrationCreate,
		ReadContext:   resourceGooglePayMerchantRegistrationRead,
		DeleteContext: resourceGooglePayMerchantRegistrationDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Google Pay Merchant Registration",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"merchant_identifier": {
				Description: "Google Pay merchant identifier",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Google Pay Merchant Registration",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Timestamp at which the Google Pay Merchant Registration was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceGooglePayMerchantRegistrationCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchantIdentifier := data.Get("merchant_identifier").(string)

	merchant, err := btClient.GooglePay.Merchant.Create(ctx, &googlepay.GooglePayMerchantRegisterRequest{
		MerchantIdentifier: &merchantIdentifier,
	})
	if err != nil {
		return apiErrorDiagnostics("Error creating Google Pay Merchant Registration:", err)
	}

	data.SetId(*merchant.ID)

	return resourceGooglePayMerchantRegistrationRead(ctx, data, meta)
}

func resourceGooglePayMerchantRegistrationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchant, err := btClient.GooglePay.Merchant.Get(ctx, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			data.SetId("")
			return nil
		}
		return apiErrorDiagnostics("Error reading Google Pay Merchant Registration:", err)
	}

	data.SetId(*merchant.ID)

	createdAt := ""
	if merchant.CreatedAt != nil {
		createdAt = merchant.CreatedAt.String()
	}

	for datumName, datumValue := range map[string]interface{}{
		"merchant_identifier": getStringValue(merchant.GetMerchantIdentifier()),
		"created_by":          getStringValue(merchant.GetCreatedBy()),
		"created_at":          createdAt,
	} {
		if err := data.Set(datumName, datumValue); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceGooglePayMerchantRegistrationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	err := btClient.GooglePay.Merchant.Delete(ctx, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}
		return apiErrorDiagnostics("Error deleting Google Pay Merchant Registration:", err)
	}

	return nil
}
