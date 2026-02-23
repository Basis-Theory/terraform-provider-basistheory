package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	basistheory "github.com/Basis-Theory/go-sdk/v5"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/Basis-Theory/go-sdk/v5/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestGooglePayMerchantRegistration(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckGooglePayMerchantRegistrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGooglePayMerchantRegistrationCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(
						"basistheory_google_pay_merchant_registration.terraform_test_google_pay_merchant", "id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttr(
						"basistheory_google_pay_merchant_registration.terraform_test_google_pay_merchant", "merchant_identifier", "terraform-test-merchant"),
				),
			},
		},
	})
}

func testAccCheckGooglePayMerchantRegistrationDestroy(s *terraform.State) error {
	btClient := basistheoryClient.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "basistheory_google_pay_merchant_registration" {
			continue
		}

		_, err := btClient.GooglePay.Merchant.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Google Pay Merchant Registration %s still exists", rs.Primary.ID)
		}

		var notFoundError *basistheory.NotFoundError
		if !errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}

const testGooglePayMerchantRegistrationCreate = `
resource "basistheory_google_pay_merchant_registration" "terraform_test_google_pay_merchant" {
	merchant_identifier = "terraform-test-merchant"
}
`
