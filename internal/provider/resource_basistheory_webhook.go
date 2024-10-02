package provider

import (
	"context"
	basistheoryV2 "github.com/Basis-Theory/go-sdk"
	basistheoryV2client "github.com/Basis-Theory/go-sdk/client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceBasisTheoryWebhook() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Unique identifier of the Webhook",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"tenant_id": {
				Description: "Tenant identifier where this Webhook was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"name": {
				Description: "Name of the Webhook",
				Type:        schema.TypeString,
				Required:    true,
			},
			"url": {
				Description: "URL of the Webhook",
				Type:        schema.TypeString,
				Required:    true,
			},
			"events": {
				Description: "List of events to subscribe to the Webhook",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created_at": {
				Description: "Timestamp at which the Webhook was created",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"created_by": {
				Description: "Identifier for who created the Webhook",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_at": {
				Description: "Timestamp at which the Webhook was last updated",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"modified_by": {
				Description: "Identifier for who last modified the Webhook",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	webhook := getWebhookFromData(data)

	request := &basistheoryV2.CreateWebhookRequest{
		Name:   webhook.Name,
		URL:    webhook.URL,
		Events: webhook.Events,
	}

	response, err := basisTheoryClient.Webhooks.Create(ctx, request)
	if err != nil {
		return apiErrorDiagnosticsV2("Error creating Webhook:", err)
	}

	data.SetId(response.ID)
	return nil
}

func resourceWebhookRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	webhook, err := basisTheoryClient.Webhooks.Get(ctx, data.Id())
	if err != nil {
		return apiErrorDiagnosticsV2("Error reading Webhook:", err)
	}

	data.SetId(webhook.ID)

	modifiedAt := ""

	if webhook.ModifiedAt != nil {
		modifiedAt = webhook.ModifiedAt.String()
	}

	for webhookDatumName, webhookDatum := range map[string]interface{}{
		"tenant_id": webhook.TenantID,
		"name":      webhook.Name,
		"url":       webhook.URL,
		"events":    webhook.Events,
		"created_at": webhook.CreatedAt.String(),
		"created_by": webhook.CreatedBy,
		"modified_at": modifiedAt,
		"modified_by": webhook.ModifiedBy,
	} {
		err := data.Set(webhookDatumName, webhookDatum)

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceWebhookUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	webhook := getWebhookFromData(data)

	request := &basistheoryV2.UpdateWebhookRequest{
		Name:   webhook.Name,
		URL:    webhook.URL,
		Events: webhook.Events,
	}

	_, err := basisTheoryClient.Webhooks.Update(ctx, data.Id(), request)
	if err != nil {
		return apiErrorDiagnosticsV2("Error updating Webhook:", err)
	}

	return nil
}

func resourceWebhookDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	basisTheoryClient := meta.(map[string]interface{})["clientV2"].(*basistheoryV2client.Client)

	err := basisTheoryClient.Webhooks.Delete(ctx, data.Id())
	if err != nil {
		return apiErrorDiagnosticsV2("Error deleting Webhook:", err)
	}

	return nil
}

func getWebhookFromData(data *schema.ResourceData) *basistheoryV2.Webhook {
	var events []string
	if dataEvents, ok := data.Get("events").(*schema.Set); ok {
		for _, event := range dataEvents.List() {
			events = append(events, event.(string))
		}
	}

	return &basistheoryV2.Webhook{
		ID: data.Id(),
		Name:   data.Get("name").(string),
		URL:    data.Get("url").(string),
		Events: events,
	}
}
