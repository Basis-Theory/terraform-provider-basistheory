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

func TestResourceReactorFormula(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckReactorFormulaDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccReactorFormulaCreate, "terraform_test_reactor_formula"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "name", "Terraform reactor formula"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "type", "private"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "description", "Terraform reactor formula"),
					resource.TestMatchResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "code", regexp.MustCompile("module.exports = async function")),
					resource.TestCheckResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "icon", "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "configuration.*", map[string]string{
							"name":        "TEST_FOO",
							"description": "foobar",
							"type":        "string",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "configuration.*", map[string]string{
							"name":        "TEST_CONFIG_BAR",
							"description": "barfoo",
							"type":        "number",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "request_parameter.*", map[string]string{
							"name":        "TEST_REQUEST_PARAM_FOO",
							"description": "foobar",
							"type":        "string",
							"optional":    "true",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "request_parameter.*", map[string]string{
							"name":        "TEST_REQUEST_PARAM_BAR",
							"description": "foobar",
							"type":        "string",
							"optional":    "false",
						}),
				),
			},
			{
				Config: fmt.Sprintf(testAccReactorFormulaUpdate, "terraform_test_reactor_formula"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "name", "Terraform reactor formula updated name"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "type", "private"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "description", "Terraform reactor formula updated description"),
					resource.TestMatchResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "code", regexp.MustCompile("const package = require")),
					resource.TestCheckResourceAttr(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "icon", "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "configuration.*", map[string]string{
							"name":        "TEST_FOO",
							"description": "foobar updated description",
							"type":        "string",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "configuration.*", map[string]string{
							"name":        "TEST_CONFIG_BAR",
							"description": "barfoo updated description",
							"type":        "number",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "request_parameter.*", map[string]string{
							"name":        "TEST_REQUEST_PARAM_FOO_UPDATED",
							"description": "foobar",
							"type":        "string",
							"optional":    "true",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_reactor_formula.terraform_test_reactor_formula", "request_parameter.*", map[string]string{
							"name":        "TEST_REQUEST_PARAM_BAR_UPDATED",
							"description": "foobar",
							"type":        "string",
							"optional":    "false",
						}),
				),
			},
		},
	})
}

const testAccReactorFormulaCreate = `
resource "basistheory_reactor_formula" "%s" {
  name = "Terraform reactor formula"
  type = "private"
  description = "Terraform reactor formula"
  code = <<-EOT
            module.exports = async function (context) {
              return context;
            };
        EOT
  icon = "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="
  configuration {
    name = "TEST_FOO"
    description = "foobar"
    type = "string"
  }
  configuration {
    name = "TEST_CONFIG_BAR"
    description = "barfoo"
    type = "number"
  }
  request_parameter {
    name = "TEST_REQUEST_PARAM_FOO"
    description = "foobar"
    type = "string"
    optional = true
  }
  request_parameter {
    name = "TEST_REQUEST_PARAM_BAR"
    description = "foobar"
    type = "string"
    optional = false
  }
}
`

const testAccReactorFormulaUpdate = `
resource "basistheory_reactor_formula" "%s" {
  name = "Terraform reactor formula updated name"
  type = "private"
  description = "Terraform reactor formula updated description"
  code = <<-EOT
			const package = require("abcd");
            module.exports = async function (context) {
              return context;
            };
        EOT
  icon = "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAAAUAAAAFCAYAAACNbyblAAAAHElEQVQI12P4//8/w38GIAXDIBKE0DHxgljNBAAO9TXL0Y4OHwAAAABJRU5ErkJggg=="
  configuration {
    name = "TEST_FOO"
    description = "foobar updated description"
    type = "string"
  }
  configuration {
    name = "TEST_CONFIG_BAR"
    description = "barfoo updated description"
    type = "number"
  }
  request_parameter {
    name = "TEST_REQUEST_PARAM_FOO_UPDATED"
    description = "foobar"
    type = "string"
    optional = true
  }
  request_parameter {
    name = "TEST_REQUEST_PARAM_BAR_UPDATED"
    description = "foobar"
    type = "string"
    optional = false
  }
}
`

func testAccCheckReactorFormulaDestroy(state *terraform.State) error {
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
		if rs.Type != "basistheory_reactor_formula" {
			continue
		}

		_, _, err := basisTheoryClient.ReactorFormulasApi.ReactorFormulasGetById(ctxWithApiKey, rs.Primary.ID).Execute()

		if !strings.Contains(err.Error(), "Not Found") {
			return err
		}
	}

	return nil
}
