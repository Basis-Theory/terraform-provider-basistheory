package provider

import (
	"context"
	"errors"

	basistheory "github.com/Basis-Theory/go-sdk/v5"
	"github.com/Basis-Theory/go-sdk/v5/applepay"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryApplePayMerchantRegistration() *schema.Resource {
	return &schema.Resource{
		Description: "Apple Pay Merchant Registration https://developers.basistheory.com/docs/api/apple-pay/api#apple-pay-merchant-registration",

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateContext: resourceApplePayMerchantRegistrationCreate,
		ReadContext:   resourceApplePayMerchantRegistrationRead,
		DeleteContext: resourceApplePayMerchantRegistrationDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Apple Pay Merchant Registration",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"merchant_identifier": {
				Description: "Apple Pay merchant identifier",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Apple Pay Merchant Registration",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Timestamp at which the Apple Pay Merchant Registration was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceApplePayMerchantRegistrationCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchantIdentifier := data.Get("merchant_identifier").(string)

	merchant, err := btClient.ApplePay.Merchant.Create(ctx, &applepay.ApplePayMerchantRegisterRequest{
		MerchantIdentifier: &merchantIdentifier,
	})
	if err != nil {
		return apiErrorDiagnostics("Error creating Apple Pay Merchant Registration:", err)
	}

	data.SetId(*merchant.ID)

	return resourceApplePayMerchantRegistrationRead(ctx, data, meta)
}

func resourceApplePayMerchantRegistrationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchant, err := btClient.ApplePay.Merchant.Get(ctx, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			data.SetId("")
			return nil
		}
		return apiErrorDiagnostics("Error reading Apple Pay Merchant Registration:", err)
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

func resourceApplePayMerchantRegistrationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	err := btClient.ApplePay.Merchant.Delete(ctx, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}
		return apiErrorDiagnostics("Error deleting Apple Pay Merchant Registration:", err)
	}

	return nil
}
