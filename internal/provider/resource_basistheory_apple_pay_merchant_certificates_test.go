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

func TestApplePayMerchantCertificates(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckApplePayMerchantCertificatesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testApplePayMerchantCertificatesCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(
						"basistheory_apple_pay_merchant_certificates.terraform_test_apple_pay_cert", "id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttrPair(
						"basistheory_apple_pay_merchant_certificates.terraform_test_apple_pay_cert", "merchant_registration_id",
						"basistheory_apple_pay_merchant_registration.terraform_test_apple_pay_merchant", "id"),
				),
			},
		},
	})
}

func testAccCheckApplePayMerchantCertificatesDestroy(s *terraform.State) error {
	btClient := basistheoryClient.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "basistheory_apple_pay_merchant_certificates" {
			continue
		}

		merchantRegistrationID := rs.Primary.Attributes["merchant_registration_id"]
		_, err := btClient.ApplePay.Merchant.Certificates.Get(context.TODO(), merchantRegistrationID, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Apple Pay Merchant Certificate %s still exists", rs.Primary.ID)
		}

		var notFoundError *basistheory.NotFoundError
		if !errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}

const testApplePayMerchantCertificatesCreate = `
resource "basistheory_apple_pay_merchant_registration" "terraform_test_apple_pay_merchant" {
	merchant_identifier = "terraform-test-apple-merchant"
}

resource "basistheory_apple_pay_merchant_certificates" "terraform_test_apple_pay_cert" {
	merchant_registration_id  = basistheory_apple_pay_merchant_registration.terraform_test_apple_pay_merchant.id
	merchant_certificate_data = "base64-encoded-cert-data"
	domain                    = "cdn.flock-dev.com"
}
`
