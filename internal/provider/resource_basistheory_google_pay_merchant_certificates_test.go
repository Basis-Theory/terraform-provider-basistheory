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

func TestGooglePayMerchantCertificates(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckGooglePayMerchantCertificatesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testGooglePayMerchantCertificatesCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(
						"basistheory_google_pay_merchant_certificates.terraform_test_google_pay_cert", "id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttrPair(
						"basistheory_google_pay_merchant_certificates.terraform_test_google_pay_cert", "merchant_registration_id",
						"basistheory_google_pay_merchant_registration.terraform_test_google_pay_merchant", "id"),
				),
			},
		},
	})
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

const testGooglePayMerchantCertificatesCreate = `
resource "basistheory_google_pay_merchant_registration" "terraform_test_google_pay_merchant" {
	merchant_identifier = "terraform-test-merchant"
}

resource "basistheory_google_pay_merchant_certificates" "terraform_test_google_pay_cert" {
	merchant_registration_id      = basistheory_google_pay_merchant_registration.terraform_test_google_pay_merchant.id
	merchant_certificate_data     = "base64-encoded-cert-data"
	merchant_certificate_password = "cert-password"
}
`
