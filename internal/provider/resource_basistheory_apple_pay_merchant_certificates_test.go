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

const (
	applePayCertResourceName = "basistheory_apple_pay_merchant_certificates.terraform_test_apple_pay_cert"
	applePayMerchantName     = "basistheory_apple_pay_merchant_registration.terraform_test_apple_pay_merchant"
)

func TestApplePayMerchantCertificates(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckApplePayMerchantCertificatesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccApplePayMerchantCertificatesConfig("cdn.flock-dev.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(applePayCertResourceName, "id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttrPair(applePayCertResourceName, "merchant_registration_id", applePayMerchantName, "id"),
					resource.TestCheckResourceAttr(applePayCertResourceName, "domain", "cdn.flock-dev.com"),
					resource.TestCheckResourceAttrSet(applePayCertResourceName, "merchant_certificate_fingerprint"),
					resource.TestCheckResourceAttrSet(applePayCertResourceName, "merchant_certificate_expiration_date"),
					resource.TestCheckResourceAttrSet(applePayCertResourceName, "payment_processor_certificate_fingerprint"),
					resource.TestCheckResourceAttrSet(applePayCertResourceName, "payment_processor_certificate_expiration_date"),
					resource.TestCheckResourceAttrSet(applePayCertResourceName, "created_by"),
					resource.TestCheckResourceAttrSet(applePayCertResourceName, "created_at"),
				),
			},
			{
				Config: testAccApplePayMerchantCertificatesConfig("cdn2.flock-dev.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(applePayCertResourceName, "id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttr(applePayCertResourceName, "domain", "cdn.flock-dev.com"),
				),
			},
		},
	})
}

func testAccApplePayMerchantCertificatesConfig(domain string) string {
	return fmt.Sprintf(`
resource "basistheory_apple_pay_merchant_registration" "terraform_test_apple_pay_merchant" {
	merchant_identifier = "terraform-test-apple-merchant"
}

resource "basistheory_apple_pay_merchant_certificates" "terraform_test_apple_pay_cert" {
	merchant_registration_id               = basistheory_apple_pay_merchant_registration.terraform_test_apple_pay_merchant.id
	merchant_certificate_data              = "%s"
	merchant_certificate_password          = "%s"
	payment_processor_certificate_data     = "%s"
	payment_processor_certificate_password = "%s"
	domain                                 = "%s"
}
`,
		os.Getenv("BT_APPLE_PAY_MERCHANT_IDENTITY_CERTIFICATE"),
		os.Getenv("BT_APPLE_PAY_MERCHANT_IDENTITY_CERTIFICATE_PASSWORD"),
		os.Getenv("BT_APPLE_PAY_PAYMENT_PROCESSING_CERTIFICATE"),
		os.Getenv("BT_APPLE_PAY_PAYMENT_PROCESSING_CERTIFICATE_PASSWORD"),
		domain,
	)
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
