package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	basistheory "github.com/Basis-Theory/go-sdk/v3"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v3/client"
	"github.com/Basis-Theory/go-sdk/v3/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceReactor(t *testing.T) {
	testAccApplicationName := "terraform_test_application_reactor_test"
	formattedTestAccApplicationCreate := fmt.Sprintf(testAccApplicationCreateWithCreateKeyTrue, testAccApplicationName)
	formattedTestAccReactorCreate := fmt.Sprintf(testAccReactorCreate, "terraform_test_reactor", testAccApplicationName)
	formattedTestAccReactorUpdate := fmt.Sprintf(testAccReactorUpdate, "terraform_test_reactor", testAccApplicationName)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckReactorDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s\n%s", formattedTestAccApplicationCreate, formattedTestAccReactorCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "name", "Terraform reactor"),
					resource.TestMatchResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "application_id", regexp.MustCompile(testUuidRegex)),
					resource.TestMatchResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "code", regexp.MustCompile("return context")),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "configuration.TEST_FOO", "TEST_FOO"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR"),
				),
			},
			{
				Config: fmt.Sprintf("%s\n%s", formattedTestAccApplicationCreate, formattedTestAccReactorUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "name", "Terraform reactor updated name"),
					resource.TestMatchResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "application_id", regexp.MustCompile(testUuidRegex)),
					resource.TestMatchResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "code", regexp.MustCompile("const package = require")),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "configuration.TEST_FOO", "TEST_FOO_UPDATED"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR_UPDATED"),
				),
			},
		},
	})
}

func TestResourceReactorWithoutApplication(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccReactorCreateWithoutApplication, "terraform_test_reactor_without_application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_without_application", "name", "Terraform reactor"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_without_application", "application_id", ""),
					resource.TestMatchResourceAttr(
						"basistheory_reactor.terraform_test_reactor_without_application", "code", regexp.MustCompile("return context")),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_without_application", "configuration.TEST_FOO", "TEST_FOO"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_without_application", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR"),
				),
			},
		},
	})
}

const testAccReactorCreate = `
resource "basistheory_reactor" "%s" {
  name = "Terraform reactor"
  application_id = "${basistheory_application.%s.id}"
  code = <<-EOT
            module.exports = async function (context) {
              return context;
            };
        EOT
  configuration = {
    TEST_FOO = "TEST_FOO"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR"
  }
}
`

const testAccReactorUpdate = `
resource "basistheory_reactor" "%s" {
  name = "Terraform reactor updated name"
  application_id = "${basistheory_application.%s.id}"
  code = <<-EOT
            const package = require("abcd");
            module.exports = async function (context) {
              return context;
            };
        EOT
  configuration = {
    TEST_FOO = "TEST_FOO_UPDATED"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR_UPDATED"
  }
}
`

const testAccReactorCreateWithoutApplication = `
resource "basistheory_reactor" "%s" {
  name = "Terraform reactor"
  code = <<-EOT
            module.exports = async function (context) {
              return context;
            };
        EOT
  configuration = {
    TEST_FOO = "TEST_FOO"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR"
  }
}
`

func testAccCheckReactorDestroy(state *terraform.State) error {
	basisTheoryClient := basistheoryClient.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "basistheory_reactor" {
			continue
		}

		_, err := basisTheoryClient.Reactors.Get(context.TODO(), rs.Primary.ID)

		var notFoundError basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}
