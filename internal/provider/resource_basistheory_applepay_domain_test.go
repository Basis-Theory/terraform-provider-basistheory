package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestApplePayDomain(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck: func() {preCheck(t) },
		ProviderFactories: getProviderFactories(),
		//CheckDestroy: testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testApplePayDomainRegister, "terraform_test_apple_pay_domain"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_applepay_domain.terraform_test_apple_pay_domain", "domain", "cdn.flock-dev.com"),
				),
			},
		},
	})
}

const testApplePayDomainRegister = `
resource "basistheory_applepay_domain" "%s" {
	domain = "cdn.flock-dev.com"
}
`