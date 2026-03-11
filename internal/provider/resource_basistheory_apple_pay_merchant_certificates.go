package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	basistheory "github.com/Basis-Theory/go-sdk/v5"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	merchantpkg "github.com/Basis-Theory/go-sdk/v5/applepay/merchant"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryApplePayMerchantCertificates() *schema.Resource {
	return &schema.Resource{
		Description: "Apple Pay Merchant Registration Certificates https://developers.basistheory.com/docs/api/apple-pay/merchant-registration",

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

		CreateContext: resourceApplePayMerchantCertificatesCreate,
		ReadContext:   resourceApplePayMerchantCertificatesRead,
		DeleteContext: resourceApplePayMerchantCertificatesDelete,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier for the Apple Pay Merchant Certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"merchant_registration_id": {
				Description: "Identifier of the Apple Pay Merchant Registration this certificate belongs to",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"merchant_certificate_data": {
				Description: "Base64-encoded PKCS#12 merchant certificate data",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"merchant_certificate_password": {
				Description: "Password for the merchant PKCS#12 certificate",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"payment_processor_certificate_data": {
				Description: "Base64-encoded PKCS#12 payment processor certificate data",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"payment_processor_certificate_password": {
				Description: "Password for the payment processor PKCS#12 certificate",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Sensitive:   true,
			},
			"domain": {
				Description: "Domain associated with this Apple Pay Merchant Certificate",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
			"payment_processor_certificate_fingerprint": {
				Description: "Fingerprint of the registered payment processor certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"payment_processor_certificate_expiration_date": {
				Description: "Expiration date of the registered payment processor certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Apple Pay Merchant Certificate",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_at": {
				Description: "Timestamp at which the Apple Pay Merchant Certificate was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceApplePayMerchantCertificatesCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchantRegistrationID := data.Get("merchant_registration_id").(string)
	certificateData := data.Get("merchant_certificate_data").(string)

	password := data.Get("merchant_certificate_password").(string)
	ppCertData := data.Get("payment_processor_certificate_data").(string)
	ppPassword := data.Get("payment_processor_certificate_password").(string)

	request := &merchantpkg.ApplePayMerchantCertificatesRegisterRequest{
		MerchantCertificateData:             &certificateData,
		MerchantCertificatePassword:         &password,
		PaymentProcessorCertificateData:     &ppCertData,
		PaymentProcessorCertificatePassword: &ppPassword,
	}

	domain := data.Get("domain").(string)
	request.Domain = &domain

	cert, err := btClient.ApplePay.Merchant.Certificates.Create(ctx, merchantRegistrationID, request)
	if err != nil {
		return apiErrorDiagnostics("Error creating Apple Pay Merchant Certificate:", err)
	}

	data.SetId(*cert.ID)

	return resourceApplePayMerchantCertificatesRead(ctx, data, meta)
}

func resourceApplePayMerchantCertificatesRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchantRegistrationID := data.Get("merchant_registration_id").(string)

	cert, err := btClient.ApplePay.Merchant.Certificates.Get(ctx, merchantRegistrationID, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			data.SetId("")
			return nil
		}
		return apiErrorDiagnostics("Error reading Apple Pay Merchant Certificate:", err)
	}

	data.SetId(*cert.ID)

	createdAt := ""
	if cert.CreatedAt != nil {
		createdAt = cert.CreatedAt.String()
	}

	merchantCertExpiration := ""
	if cert.MerchantCertificateExpirationDate != nil {
		merchantCertExpiration = cert.MerchantCertificateExpirationDate.String()
	}

	ppCertExpiration := ""
	if cert.PaymentProcessorCertificateExpirationDate != nil {
		ppCertExpiration = cert.PaymentProcessorCertificateExpirationDate.String()
	}

	for datumName, datumValue := range map[string]interface{}{
		"domain":                                        getStringValue(cert.GetDomain()),
		"merchant_certificate_fingerprint":              getStringValue(cert.GetMerchantCertificateFingerprint()),
		"merchant_certificate_expiration_date":          merchantCertExpiration,
		"payment_processor_certificate_fingerprint":     getStringValue(cert.GetPaymentProcessorCertificateFingerprint()),
		"payment_processor_certificate_expiration_date": ppCertExpiration,
		"created_by":                                    getStringValue(cert.GetCreatedBy()),
		"created_at":                                    createdAt,
	} {
		if err := data.Set(datumName, datumValue); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceApplePayMerchantCertificatesDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	btClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	merchantRegistrationID := data.Get("merchant_registration_id").(string)

	err := btClient.ApplePay.Merchant.Certificates.Delete(ctx, merchantRegistrationID, data.Id())
	if err != nil {
		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return nil
		}
		return apiErrorDiagnostics("Error deleting Apple Pay Merchant Certificate:", err)
	}

	return nil
}
