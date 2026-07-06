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
	googlePayCertResourceName = "basistheory_google_pay_merchant_certificates.terraform_test_google_pay_cert"
	googlePayMerchantName     = "basistheory_google_pay_merchant_registration.terraform_test_google_pay_merchant"
)

func TestGooglePayMerchantCertificates(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckGooglePayMerchantCertificatesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGooglePayMerchantCertificatesConfig("terraform-test-google-merchant"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(googlePayCertResourceName, "id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttrPair(googlePayCertResourceName, "merchant_registration_id", googlePayMerchantName, "id"),
					resource.TestCheckResourceAttrSet(googlePayCertResourceName, "merchant_certificate_fingerprint"),
					resource.TestCheckResourceAttrSet(googlePayCertResourceName, "merchant_certificate_expiration_date"),
					resource.TestCheckResourceAttrSet(googlePayCertResourceName, "created_by"),
					resource.TestCheckResourceAttrSet(googlePayCertResourceName, "created_at"),
				),
			},
			{
				Config: testAccGooglePayMerchantCertificatesConfig("terraform-test-google-merchant-2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(googlePayCertResourceName, "id", regexp.MustCompile(testUuidRegex)),
				),
			},
		},
	})
}

func testAccGooglePayMerchantCertificatesConfig(merchantIdentifier string) string {
	return fmt.Sprintf(`
resource "basistheory_google_pay_merchant_registration" "terraform_test_google_pay_merchant" {
	merchant_identifier = "%s"
}

resource "basistheory_google_pay_merchant_certificates" "terraform_test_google_pay_cert" {
	merchant_registration_id      = basistheory_google_pay_merchant_registration.terraform_test_google_pay_merchant.id
	merchant_certificate_data     = "%s"
	merchant_certificate_password = "%s"
}
`,
		merchantIdentifier,
		os.Getenv("BT_GOOGLE_PAY_MERCHANT_IDENTITY_CERTIFICATE"),
		os.Getenv("BT_GOOGLE_PAY_MERCHANT_IDENTITY_CERTIFICATE_PASSWORD"),
	)
}

func testAccCheckGooglePayMerchantCertificatesDestroy(s *terraform.State) error {
	btClient := basistheoryClient.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "basistheory_google_pay_merchant_certificates" {
			continue
		}

		merchantRegistrationID := rs.Primary.Attributes["merchant_registration_id"]
		_, err := btClient.GooglePay.Merchant.Certificates.Get(context.TODO(), merchantRegistrationID, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Google Pay Merchant Certificate %s still exists", rs.Primary.ID)
		}

		var notFoundError *basistheory.NotFoundError
		if !errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}
