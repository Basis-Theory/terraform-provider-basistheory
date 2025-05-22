package provider

import (
	"context"
	"errors"
	"fmt"
	basistheory "github.com/Basis-Theory/go-sdk/v2"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v2/client"
	"github.com/Basis-Theory/go-sdk/v2/option"
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
						"basistheory_webhook.terraform_test_webhook", "url", "https://echo.flock-dev.com/terraform-webhook"),
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "notify_email", "here@there.com"),
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "events.0", "token.created"),
					pauseForSeconds(2), // Required to avoid error `The webhook subscription is undergoing another concurrent operation. Please wait a few seconds, then try again.
				),
			},
			{
				Config: fmt.Sprintf(testWebhookUpdate, "terraform_test_webhook"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "name", "(Deletable) Terraform Webhook updated"),
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "url", "https://echo.flock-dev.com/terraform-webhook-updated"),
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "notify_email", "here@somewhere-else.com"),
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "events.0", "token.created"),
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "events.1", "token.updated"),
					pauseForSeconds(2), // Required to avoid error `The webhook subscription is undergoing another concurrent operation. Please wait a few seconds, then try again.
				),
			},
		},
	})
}

func TestResourceWebhook_UpdateOptionalAttributesFromNilToSomething(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy: testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildWebhookWithOptionalParameters("terraform_test_webhook", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "notify_email"),
					pauseForSeconds(2), // Required to avoid error `The webhook subscription is undergoing another concurrent operation. Please wait a few seconds, then try again.
				),
			},
			{
				Config: buildWebhookWithOptionalParameters("terraform_test_webhook", "notify_email = \"here@somewhere-else.com\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "notify_email", "here@somewhere-else.com"),
					pauseForSeconds(2), // Required to avoid error `The webhook subscription is undergoing another concurrent operation. Please wait a few seconds, then try again.
				),
			},
		},
	})
}

func TestResourceWebhook_UpdateOptionalAttributesFromSomethingToNil(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy: testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildWebhookWithOptionalParameters("terraform_test_webhook", "notify_email = \"here@there.com\""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "notify_email", "here@there.com"),
					pauseForSeconds(2), // Required to avoid error `The webhook subscription is undergoing another concurrent operation. Please wait a few seconds, then try again.
				),
			},
			{
				Config: buildWebhookWithOptionalParameters("terraform_test_webhook", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_webhook.terraform_test_webhook", "notify_email", ""),
					pauseForSeconds(2), // Required to avoid error `The webhook subscription is undergoing another concurrent operation. Please wait a few seconds, then try again.
				),
			},
		},
	})
}

func testAccCheckWebhookDestroy(state *terraform.State) error {
	basisTheoryClient := basistheoryClient.NewClient(
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

func buildWebhookWithOptionalParameters(resourceName string, opts string) string {
	return fmt.Sprintf(`
resource "basistheory_webhook" "%s" {
	name = "(Deletable) Terraform Webhook"
	url = "https://echo.flock-dev.com/terraform-webhook"
	%s
	events = ["token.created"]
}
`, resourceName, opts)
}

const testWebhookCreate = `
resource "basistheory_webhook" "%s" {
	name = "(Deletable) Terraform Webhook"
	url = "https://echo.flock-dev.com/terraform-webhook"
	notify_email = "here@there.com"
	events = ["token.created"]
}
`

const testWebhookUpdate = `
resource "basistheory_webhook" "%s" {
	name = "(Deletable) Terraform Webhook updated"
	url = "https://echo.flock-dev.com/terraform-webhook-updated"
	notify_email = "here@somewhere-else.com"
	events = ["token.created", "token.updated"]
}
`
