package provider

import (
	"context"
	"errors"
	"fmt"
	basistheory "github.com/Basis-Theory/go-sdk"
	basistheoryV2 "github.com/Basis-Theory/go-sdk/client"
	"github.com/Basis-Theory/go-sdk/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
	"time"
)

func TestResourceWebhook(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy: testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testWebhookCreate, "terraform_test_webhook"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "name", "(Deletable) Terraform Webhook"),
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "url", "https://echo.basistheory.com/terraform-webhook"),
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "events.0", "token.created"),
					pauseForSeconds(2), // Required to avoid error `The webhook subscription is undergoing another concurrent operation. Please wait a few seconds, then try again.
				),
			},
		},
	})
}

func testAccCheckWebhookDestroy(state *terraform.State) error {
	basisTheoryClient := basistheoryV2.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "basistheory_webhook" {
			continue
		}

		_, err := basisTheoryClient.Webhooks.Get(context.TODO(), rs.Primary.ID)

		var notFoundError basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}

func pauseForSeconds(duration time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(duration * time.Second)
		return nil
	}
}

const testWebhookCreate = `
resource "basistheory_webhook" "%s" {
	name = "(Deletable) Terraform Webhook"
	url = "https://echo.basistheory.com/terraform-webhook"
	events = ["token.created"]
}
`
