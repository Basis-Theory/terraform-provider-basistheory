package provider

import (
	"context"
	"errors"
	"fmt"
	basistheory "github.com/Basis-Theory/go-sdk/v5"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/Basis-Theory/go-sdk/v5/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"regexp"
	"testing"
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

func TestResourceApplicationKey_HandlesGraceful404(t *testing.T) {
	appName := "terraform_test_application_key_404"
	config := fmt.Sprintf("%s\n%s",
		fmt.Sprintf(testAccApplicationCreate, appName),
		fmt.Sprintf(testAccApplicationKeyCreate, "terraform_test_application_key_404", appName),
	)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("basistheory_application_key.terraform_test_application_key_404", "id"),
					deleteApplicationKeyExternally("basistheory_application_key.terraform_test_application_key_404"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func deleteApplicationKeyExternally(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client := basistheoryClient.NewClient(
			option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
			option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
		)

		applicationId := rs.Primary.Attributes["application_id"]
		return client.ApplicationKeys.Delete(context.TODO(), applicationId, rs.Primary.ID)
	}
}

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

		var notFoundError *basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}
