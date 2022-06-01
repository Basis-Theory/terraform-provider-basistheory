package provider

import (
	"context"
	"errors"
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

func TestResourceApplication(t *testing.T) {
	testAccApplicationKey := ""

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccApplicationCreate, "terraform_test_application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "name", "Terraform application"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "type", "server_to_server"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.0", "token:general:read:low"),
					testAccSetApplicationKeyAfterCreate(&testAccApplicationKey),
				),
			},
			{
				Config: fmt.Sprintf(testAccApplicationUpdate, "terraform_test_application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "name", "Terraform application updated name"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "type", "server_to_server"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.0", "token:bank:read:low"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.1", "token:general:read:moderate"),
					testAccCheckApplicationKeyHasNotChangedBetweenOperations(&testAccApplicationKey),
				),
			},
		},
	})
}

func TestResourceApplication_invalid_permission(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationCreateWithInvalidPermission,
				ExpectError: regexp.MustCompile(`(?s)Error creating Application:.*Status Code: 400.*Title:.*Errors:.*`),
			},
		},
	})
}

func TestResourceApplication_invalid_type(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationCreateWithInvalidType,
				ExpectError: regexp.MustCompile(`expected type to be one of \[elements public server_to_server management], got foo`),
			},
		},
	})
}

const testAccApplicationCreate = `
resource "basistheory_application" "%s" {
  name = "Terraform application"
  type = "server_to_server"
  permissions = ["token:general:read:low"]
}
`

const testAccApplicationCreateWithInvalidPermission = `
resource "basistheory_application" "terraform_test_application" {
  name = "Terraform application"
  type = "server_to_server"
  permissions = ["token:general:read:foo"]
}
`

const testAccApplicationCreateWithInvalidType = `
resource "basistheory_application" "terraform_test_application" {
  name = "Terraform application"
  type = "foo"
  permissions = ["token:general:read:low"]
}
`

const testAccApplicationUpdate = `
resource "basistheory_application" "%s" {
  name = "Terraform application updated name"
  type = "server_to_server"
  permissions = ["token:general:read:moderate", "token:bank:read:low"]
}
`

func testAccCheckApplicationDestroy(state *terraform.State) error {
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
		if rs.Type != "basistheory_application" {
			continue
		}

		_, _, err := basisTheoryClient.ApplicationsApi.ApplicationsGetById(ctxWithApiKey, rs.Primary.ID).Execute()

		if !strings.Contains(err.Error(), "Not Found") {
			return err
		}
	}

	return nil
}

func testAccSetApplicationKeyAfterCreate(testAccApplicationKey *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "basistheory_application" {
				continue
			}

			*testAccApplicationKey = rs.Primary.Attributes["key"]
		}

		return nil
	}
}

func testAccCheckApplicationKeyHasNotChangedBetweenOperations(appKeyFromCreate *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		for _, rs := range state.RootModule().Resources {
			if rs.Type != "basistheory_application" {
				continue
			}

			stateAppKey := rs.Primary.Attributes["key"]

			if stateAppKey != *appKeyFromCreate {
				return errors.New("application key from create is different from application key after update")
			}
		}

		return nil
	}
}
