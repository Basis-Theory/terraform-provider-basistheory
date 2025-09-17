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

func TestResourceApplicationKey(t *testing.T) {
	testAccApplicationName := "terraform_test_application_applicationkey_test"
	formattedTestAccApplicationCreate := fmt.Sprintf(testAccApplicationCreate, testAccApplicationName)
	formattedTestAccApplicationKeyCreate := fmt.Sprintf(testAccApplicationKeyCreate, "terraform_test_application_key", testAccApplicationName)
	formattedTestAccApplicationKeyUpdate := fmt.Sprintf(testAccApplicationKeyUpdate, "terraform_test_application_key")

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckApplicationKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s\n%s", formattedTestAccApplicationCreate, formattedTestAccApplicationKeyCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(
						"basistheory_application_key.terraform_test_application_key", "key", regexp.MustCompile(testApplicationKeyRegex)),
					resource.TestMatchResourceAttr(
						"basistheory_application_key.terraform_test_application_key", "application_id", regexp.MustCompile(testUuidRegex)),
				),
			},
			{
				Config:      fmt.Sprintf("%s\n%s", formattedTestAccApplicationCreate, formattedTestAccApplicationKeyUpdate),
				ExpectError: regexp.MustCompile(`Updating ApplicationKey is not supported`),
			},
		},
	})
}

const testAccApplicationKeyCreate = `
resource "basistheory_application_key" "%s" {
  application_id = "${basistheory_application.%s.id}"
}
`

const testAccApplicationKeyUpdate = `
resource "basistheory_application_key" "%s" {
  application_id = "310094F9-299D-423D-9144-3DF17571F63E" # non-existent application
}
`

func testAccCheckApplicationKeyDestroy(state *terraform.State) error {
	basisTheoryClient := basistheoryClient.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "basistheory_application_key" {
			continue
		}

		applicationId := rs.Primary.Attributes["application_id"]
		keyId := rs.Primary.ID
		_, err := basisTheoryClient.ApplicationKeys.Get(context.TODO(), applicationId, keyId)

		var notFoundError basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}
