package provider

import (
	"context"
	basistheory "github.com/Basis-Theory/go-sdk"
	"github.com/Basis-Theory/go-sdk/applepay"
	basistheoryClient "github.com/Basis-Theory/go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceApplePayDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplePayDomainCreate,
		ReadContext: resourceApplePayDomainRead,
		UpdateContext: resourceApplePayDomainCreate,
		DeleteContext: resourceApplePayDomainDelete,
		Schema: map[string]*schema.Schema{
			"domain": {
				Description: "Public domain of hosted application",
				Type:        schema.TypeString,
				Required:    true,
			},
		},
	}
}

func resourceApplePayDomainCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	request := &applepay.ApplePayDomainRegistrationRequest{
		Domain: data.Get("domain").(string),
	}

	_, err := basisTheoryClient.ApplePay.Domain.Register(ctx, request)
	if err != nil {
		return apiErrorDiagnostics("Error registering Apple Pay domain:", err)
	}

	data.SetId(request.Domain)
	return nil
}

func resourceApplePayDomainRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	response, err := basisTheoryClient.ApplePay.Domain.Get(ctx)
	if err != nil {
		return apiErrorDiagnostics("Error reading Apple Pay domains:", err)
	}

	domains := response.GetDomains()
	if domains == nil {
		return apiErrorDiagnostics("No Apple Pay domains retrieved:", nil)
	}

	var matchedDomain *basistheory.DomainRegistrationResponse
	for _, domain := range domains {
		if domain.Domain != nil && *domain.Domain == data.Id() {
			matchedDomain = domain
			break
		}
	}

	if matchedDomain == nil {
		return diag.Errorf("No domain found with ID: %s", data.Id())
	}

	return nil
}

func resourceApplePayDomainDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	request := &applepay.ApplePayDomainDeregistrationRequest{
		Domain: data.Id(),
	}

	err := basisTheoryClient.ApplePay.Domain.Deregister(ctx, request)
	if err != nil {
		return apiErrorDiagnostics("Error deregistering Apple Pay domain:", err)
	}

	return nil
}