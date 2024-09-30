package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestResourceWebhook(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		//CheckDestroy:
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testWebhookCreate, "terraform_test_webhook"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_webhook", "name", "(Deletable) Terraform Webhook"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_webhook", "url", "https://echo.basistheory.com/terraform-webhook"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_webhook", "events.0", "token.created"),
				),
			},
		},
	})
}

const testWebhookCreate = `
resource "basistheory_webhook" "%s" {
	name = "(Deletable) Terraform Webhook"
	url = "https://echo.basistheory.com/terraform-webhook"
	events = ["token.created"]
}
`
