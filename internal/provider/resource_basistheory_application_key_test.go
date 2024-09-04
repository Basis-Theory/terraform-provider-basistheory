package provider

import (
	"context"
	"fmt"
	"github.com/Basis-Theory/basistheory-go/v6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"regexp"
	"strconv"
	"strings"
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

func testAccCheckApplicationKeyDestroy(state *terraform.State) error {
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
		if rs.Type != "basistheory_application_key" {
			continue
		}

		applicationId := rs.Primary.Attributes["application_id"]
		keyId := rs.Primary.ID
		_, _, err := basisTheoryClient.ApplicationKeysApi.GetById(ctxWithApiKey, applicationId, keyId).Execute()

		if !strings.Contains(err.Error(), "Not Found") {
			return err
		}
	}

	return nil
}
