package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	basistheoryClient "github.com/Basis-Theory/go-sdk/v3/client"
	"github.com/Basis-Theory/go-sdk/v3/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestApplePayDomainMultiple(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckApplePayDomainDestroy,
		Steps: []resource.TestStep{
			{
				Config: testApplePayDomainRegisterCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_applepay_domain.terraform_test_apple_pay_domain", "domains.0", "cdn.flock-dev.com"),
				),
			},
			{
				Config: testApplePayDomainRegisterUpdateToMany,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_applepay_domain.terraform_test_apple_pay_domain", "domains.#", "2"), // Counting entries is good enough
				),
			},
			{
				Config: testApplePayDomainRegisterDeleteFirst,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_applepay_domain.terraform_test_apple_pay_domain", "domains.0", "cdn.basistheory.com"),
				),
			},
		},
	})
}

func testAccCheckApplePayDomainDestroy(s *terraform.State) error {
	basisTheoryClient := basistheoryClient.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	response, err := basisTheoryClient.ApplePay.Domain.Get(context.TODO())
	if err != nil {
		return err
	}

	domains := response.GetDomains()
	if domains != nil && len(domains) != 0 {
		return fmt.Errorf("Unexpected Apple Pay domains found: %+v", domains)
	}

	return nil
}

const testApplePayDomainRegisterCreate = `
resource "basistheory_applepay_domain" "terraform_test_apple_pay_domain" {
	domains = ["cdn.flock-dev.com" ]
}
`

const testApplePayDomainRegisterUpdateToMany = `
resource "basistheory_applepay_domain" "terraform_test_apple_pay_domain" {
	domains = ["cdn.flock-dev.com", "cdn.basistheory.com" ]
}
`

const testApplePayDomainRegisterDeleteFirst = `
resource "basistheory_applepay_domain" "terraform_test_apple_pay_domain" {
	domains = ["cdn.basistheory.com" ]
}
`
