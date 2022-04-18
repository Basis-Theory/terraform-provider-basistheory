package provider

import (
	"context"
	"fmt"
	"github.com/Basis-Theory/basistheory-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestResourceReactor(t *testing.T) {
	formattedTestAccCreateReactorFormulaCreate := fmt.Sprintf(testAccReactorFormulaCreate, "terraform_test_reactor_formula_react_test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckReactorDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s\n%s", formattedTestAccCreateReactorFormulaCreate, testAccReactorCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "name", "Terraform reactor"),
					resource.TestMatchResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "formula_id", regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "configuration.TEST_FOO", "TEST_FOO"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR"),
				),
			},
			{
				Config: fmt.Sprintf("%s\n%s", formattedTestAccCreateReactorFormulaCreate, testAccReactorUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "name", "Terraform reactor updated name"),
					resource.TestMatchResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "formula_id", regexp.MustCompile("^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$")),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "configuration.TEST_FOO", "TEST_FOO_UPDATED"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR_UPDATED"),
				),
			},
		},
	})
}

const testAccReactorCreate = `
resource "basistheory_reactor" "terraform_test_reactor" {
  name = "Terraform reactor"
  formula_id = "${basistheory_reactor_formula.terraform_test_reactor_formula_react_test.id}"
  configuration = {
    TEST_FOO = "TEST_FOO"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR"
  }
}
`
const testAccReactorUpdate = `
resource "basistheory_reactor" "terraform_test_reactor" {
  name = "Terraform reactor updated name"
  formula_id = "${basistheory_reactor_formula.terraform_test_reactor_formula_react_test.id}"
  configuration = {
    TEST_FOO = "TEST_FOO_UPDATED"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR_UPDATED"
  }
}
`

func testAccCheckReactorDestroy(state *terraform.State) error {
	ctxWithApiKey := context.WithValue(context.Background(), basistheory.ContextAPIKeys, map[string]basistheory.APIKey{
		"ApiKey": {Key: os.Getenv("BASISTHEORY_API_KEY")},
	})

	urlArray := strings.Split(os.Getenv("BASISTHEORY_API_URL"), "://")
	configuration := basistheory.NewConfiguration()
	configuration.Scheme = urlArray[0]
	configuration.Host = urlArray[1]
	configuration.DefaultHeader = map[string]string{
		"Keep-Alive": strconv.Itoa(5),
	}
	basisTheoryClient := basistheory.NewAPIClient(configuration)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "basistheory_reactor" {
			continue
		}

		_, _, err := basisTheoryClient.ReactorsApi.ReactorGetById(ctxWithApiKey, rs.Primary.ID).Execute()

		if !strings.Contains(err.Error(), "Not Found") {
			return err
		}
	}

	return nil
}
