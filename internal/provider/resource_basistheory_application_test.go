package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/Basis-Theory/basistheory-go/v5"
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
						"basistheory_application.terraform_test_application", "type", "private"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.0", "token:read"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "create_key", "false"),
				),
			},
			{
				Config: fmt.Sprintf(testAccApplicationUpdate, "terraform_test_application", "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "name", "Terraform application updated name"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "type", "private"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.0", "token:read"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.1", "token:search"),
					testAccCheckApplicationKeyHasNotChangedBetweenOperations(&testAccApplicationKey),
				),
			},
		},
	})
}

func TestResourceApplication_with_create_key_true(t *testing.T) {
	testAccApplicationKey := ""

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccApplicationCreateWithCreateKeyTrue, "terraform_test_application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "name", "Terraform application"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "type", "private"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.0", "token:read"),
					testAccSetApplicationKeyAfterCreate(&testAccApplicationKey),
				),
			},
			{
				Config: fmt.Sprintf(testAccApplicationUpdate, "terraform_test_application", "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "name", "Terraform application updated name"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "type", "private"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.0", "token:read"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.1", "token:search"),
					testAccCheckApplicationKeyHasNotChangedBetweenOperations(&testAccApplicationKey),
				),
			},
		},
	})
}

func TestResourceApplication_with_create_key_true_and_updating_create_key(t *testing.T) {
	testAccApplicationKey := ""

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccApplicationCreateWithCreateKeyTrue, "terraform_test_application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "name", "Terraform application"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "type", "private"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "permissions.0", "token:read"),
					testAccSetApplicationKeyAfterCreate(&testAccApplicationKey),
				),
			},
			{
				Config:      fmt.Sprintf(testAccApplicationUpdate, "terraform_test_application", "false"),
				ExpectError: regexp.MustCompile(`Updating 'create_key' is not supported`),
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
				ExpectError: regexp.MustCompile(`expected type to be one of \[public private management], got foo`),
			},
		},
	})
}

func TestResourceApplication_with_access_rules(t *testing.T) {
	testAccApplicationKey := ""

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccApplicationWithAccessRulesCreate, "terraform_test_application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "name", "Terraform application"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "type", "private"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_application.terraform_test_application", "rule.*", map[string]string{
							"description":   "Test rule",
							"priority":      "1",
							"container":     "/",
							"transform":     "mask",
							"permissions.0": "token:read",
						}),
					testAccSetApplicationKeyAfterCreate(&testAccApplicationKey),
				),
			},
			{
				Config: fmt.Sprintf(testAccApplicationWithAccessRulesUpdate, "terraform_test_application"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "name", "Terraform application updated name"),
					resource.TestCheckResourceAttr(
						"basistheory_application.terraform_test_application", "type", "private"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_application.terraform_test_application", "rule.*", map[string]string{
							"description":   "Updated test rule",
							"priority":      "1",
							"container":     "/general/",
							"transform":     "reveal",
							"permissions.0": "token:read",
							"permissions.1": "token:search",
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"basistheory_application.terraform_test_application", "rule.*", map[string]string{
							"description":   "New rule",
							"priority":      "2",
							"container":     "/pci/",
							"transform":     "redact",
							"permissions.0": "token:use",
						}),
					testAccCheckApplicationKeyHasNotChangedBetweenOperations(&testAccApplicationKey),
				),
			},
		},
	})
}

func TestResourceApplication_with_access_rules_having_invalid_permission(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationWithAccessRulesCreateWithInvalidPermission,
				ExpectError: regexp.MustCompile(`(?s)Error creating Application:.*Status Code: 400.*Title:.*Errors:.*`),
			},
		},
	})
}

func TestResourceApplication_with_access_rules_having_invalid_transform(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccApplicationWithAccessRulesCreateWithInvalidTransform,
				ExpectError: regexp.MustCompile(`expected rule.0.transform to be one of \[mask redact reveal], got foo`),
			},
		},
	})
}

const testAccApplicationCreate = `
resource "basistheory_application" "%s" {
  name = "Terraform application"
  type = "private"
  permissions = ["token:read"]
}
`

const testAccApplicationCreateWithCreateKeyTrue = `
resource "basistheory_application" "%s" {
  name = "Terraform application"
  type = "private"
  permissions = ["token:read"]
  create_key = true
}
`

const testAccApplicationCreateWithInvalidPermission = `
resource "basistheory_application" "terraform_test_application" {
  name = "Terraform application"
  type = "private"
  permissions = ["token:foo"]
}
`

const testAccApplicationCreateWithInvalidType = `
resource "basistheory_application" "terraform_test_application" {
  name = "Terraform application"
  type = "foo"
  permissions = ["token:read"]
}
`

const testAccApplicationUpdate = `
resource "basistheory_application" "%s" {
  name = "Terraform application updated name"
  type = "private"
  permissions = ["token:read", "token:search"]
  create_key = %s
}
`

const testAccApplicationWithAccessRulesCreate = `
resource "basistheory_application" "%s" {
  name = "Terraform application"
  type = "private"
  rule {
	description = "Test rule"
	priority = 1
	container = "/"
	transform = "mask"
	permissions = ["token:read"]
  }
}
`

const testAccApplicationWithAccessRulesCreateWithInvalidPermission = `
resource "basistheory_application" "terraform_test_application" {
  name = "Terraform application"
  type = "private"
  rule {
	description = "Test rule"
	priority = 1
	container = "/"
	transform = "mask"
	permissions = ["token:foo"]
  }
}
`

const testAccApplicationWithAccessRulesCreateWithInvalidTransform = `
resource "basistheory_application" "terraform_test_application" {
  name = "Terraform application"
  type = "private"
  rule {
	description = "Test rule"
	priority = 1
	container = "/"
	transform = "foo"
	permissions = ["token:read"]
  }
}
`

const testAccApplicationWithAccessRulesUpdate = `
resource "basistheory_application" "%s" {
  name = "Terraform application updated name"
  type = "private"
  rule {
	description = "Updated test rule"
	priority = 1
	container = "/general/"
	transform = "reveal"
	permissions = ["token:read", "token:search"]
  }
  rule {
	description = "New rule"
	priority = 2
	container = "/pci/"
	transform = "redact"
	permissions = ["token:use"]
  }
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

		_, _, err := basisTheoryClient.ApplicationsApi.GetById(ctxWithApiKey, rs.Primary.ID).Execute()

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
