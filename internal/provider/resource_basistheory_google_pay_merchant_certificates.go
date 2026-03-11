package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	basistheory "github.com/Basis-Theory/go-sdk/v5"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	merchantpkg "github.com/Basis-Theory/go-sdk/v5/googlepay/merchant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryGooglePayMerchantCertificates() *schema.Resource {
	return &schema.Resource{
		Description: "Google Pay Merchant Registration Certificates https://developers.basistheory.com/docs/api/google-pay/merchant-registration",

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				parts := strings.SplitN(data.Id(), "/", 2)
				if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
					return nil, fmt.Errorf("invalid import ID %q, expected {merchant_registration_id}/{certificate_id}", data.Id())
				}
				if err := data.Set("merchant_registration_id", parts[0]); err != nil {
					return nil, err
				}
				data.SetId(parts[1])
				return []*schema.ResourceData{data}, nil
			},
		},

		CreateContext: resourceGooglePayMerchantCertificatesCreate,
		ReadContext:   resourceGooglePayMerchantCertificatesRead,
		DeleteContext: resourceGooglePayMerchantCertificatesDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Google Pay Merchant Certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"merchant_registration_id": {
				Description: "Identifier of the Google Pay Merchant Registration this certificate belongs to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"merchant_certificate_data": {
				Description: "Base64-encoded PKCS#12 certificate data",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"merchant_certificate_password": {
				Description: "Password for the PKCS#12 certificate",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"merchant_certificate_fingerprint": {
				Description: "Fingerprint of the registered merchant certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"merchant_certificate_expiration_date": {
				Description: "Expiration date of the registered merchant certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Google Pay Merchant Certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Timestamp at which the Google Pay Merchant Certificate was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceGooglePayMerchantCertificatesCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchantRegistrationID := data.Get("merchant_registration_id").(string)
	certificateData := data.Get("merchant_certificate_data").(string)

	password := data.Get("merchant_certificate_password").(string)

	request := &merchantpkg.GooglePayMerchantCertificatesRegisterRequest{
		MerchantCertificateData:     &certificateData,
		MerchantCertificatePassword: &password,
	}

	cert, err := btClient.GooglePay.Merchant.Certificates.Create(ctx, merchantRegistrationID, request)
	if err != nil {
		return apiErrorDiagnostics("Error creating Google Pay Merchant Certificate:", err)
	}

	data.SetId(*cert.ID)

	return resourceGooglePayMerchantCertificatesRead(ctx, data, meta)
}

func resourceGooglePayMerchantCertificatesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchantRegistrationID := data.Get("merchant_registration_id").(string)

	cert, err := btClient.GooglePay.Merchant.Certificates.Get(ctx, merchantRegistrationID, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			data.SetId("")
			return nil
		}
		return apiErrorDiagnostics("Error reading Google Pay Merchant Certificate:", err)
	}

	data.SetId(*cert.ID)

	createdAt := ""
	if cert.CreatedAt != nil {
		createdAt = cert.CreatedAt.String()
	}

	expirationDate := ""
	if cert.MerchantCertificateExpirationDate != nil {
		expirationDate = cert.MerchantCertificateExpirationDate.String()
	}

	for datumName, datumValue := range map[string]interface{}{
		"merchant_certificate_fingerprint":    getStringValue(cert.GetMerchantCertificateFingerprint()),
		"merchant_certificate_expiration_date": expirationDate,
		"created_by":                           getStringValue(cert.GetCreatedBy()),
		"created_at":                           createdAt,
	} {
		if err := data.Set(datumName, datumValue); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceGooglePayMerchantCertificatesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchantRegistrationID := data.Get("merchant_registration_id").(string)

	err := btClient.GooglePay.Merchant.Certificates.Delete(ctx, merchantRegistrationID, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}
		return apiErrorDiagnostics("Error deleting Google Pay Merchant Certificate:", err)
	}

	return nil
}
