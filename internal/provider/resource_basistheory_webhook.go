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

	response, err := basisTheoryClient.Webhooks.Create(context.TODO(), request)
	if err != nil {
		return apiErrorDiagnosticsV2("Error creating Webhook:", err)
	}

	data.SetId(response.ID)
	return nil
}

func resourceWebhookRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceWebhookUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceWebhookDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
