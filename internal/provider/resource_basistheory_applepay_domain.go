package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/Basis-Theory/go-sdk/applepay"
	basistheoryClient "github.com/Basis-Theory/go-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"net/http"
)

func resourceApplePayDomain() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceApplePayDomainCreate,
		ReadContext: resourceApplePayDomainRead,
		UpdateContext: resourceApplePayDomainCreate,
		DeleteContext: resourceApplePayDomainDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Internal identifier for Apple Pay domain registrations",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"domains": {
				Description: "Public domains of hosted applications",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceApplePayDomainCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	var domains []string
	if dataEvents, ok := data.Get("domains").(*schema.Set); ok {
		for _, domain := range dataEvents.List() {
			domains = append(domains, domain.(string))
		}
	}
	request := &applepay.ApplePayDomainRegistrationListRequest{
		Domains: domains,
	}

	_, err := basisTheoryClient.ApplePay.Domain.RegisterAll(ctx, request)
	if err != nil {
		return apiErrorDiagnostics("Error registering Apple Pay domain:", err)
	}

	data.SetId("applepayDomains")
	return nil
}

func resourceApplePayDomainRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["client"].(*basistheoryClient.Client)

	response, err := basisTheoryClient.ApplePay.Domain.Get(ctx)
	if err != nil {
		return apiErrorDiagnostics("Error reading Apple Pay domains:", err)
	}

	data.SetId("applepayDomains")

	domains := response.GetDomains()
	if domains == nil {
		return apiErrorDiagnostics("No Apple Pay domains retrieved:", nil)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceApplePayDomainDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This function calls the raw HTTP API because the `go-sdk` defines the ApplePayDomainRegistrationListRequest
	// with "omitempty" for the "Domains" attribute.
	// That results in an HTTP 400 since the Domains attribute is not included in the HTTP call. :/

	// 	type ApplePayDomainRegistrationListRequest struct {
	//		Domains []string `json:"domains,omitempty" url:"-"`
	//	}


	body := map[string]interface{}{
		"domains": []string{},
	}

	// Marshaling the request body to JSON
	jsonData, err := json.Marshal(body)
	if err != nil {
		return apiErrorDiagnostics("Error deregistering Apple Pay domains:", err)
	}

	// Create a new HTTP request
	url := meta.(map[string]interface{})["api_url"].(string)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, url + "/connections/apple-pay/domain-registration", bytes.NewBuffer(jsonData))
	if err != nil {
		return apiErrorDiagnostics("Error deregistering Apple Pay domains:", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("BT-API-KEY", meta.(map[string]interface{})["api_key"].(string))

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return apiErrorDiagnostics("Error deregistering Apple Pay domains:", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusOK {
		return nil
	} else {
		return apiErrorDiagnostics("Error deregistering Apple Pay domains:", err)
	}

	return nil
}