package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	basistheory "github.com/Basis-Theory/go-sdk"
	basistheoryClient "github.com/Basis-Theory/go-sdk/client"
	"github.com/Basis-Theory/go-sdk/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceClientEncryptionKey(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckClientEncryptionKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccClientEncryptionKeyCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("basistheory_client_encryption_key.test", "id"),
					resource.TestCheckResourceAttrSet("basistheory_client_encryption_key.test", "expires_at"),
				),
			},
			{
				Config: testAccClientEncryptionKeyCreateWithoutExpiresAt,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("basistheory_client_encryption_key.test", "id"),
					resource.TestCheckResourceAttrSet("basistheory_client_encryption_key.test", "expires_at"),
				),
			},
			{
				ResourceName:      "basistheory_client_encryption_key.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var testAccClientEncryptionKeyCreate = `
resource "basistheory_client_encryption_key" "test" {
  expires_at = "%s"
}
`

var testAccClientEncryptionKeyCreateWithoutExpiresAt = `
resource "basistheory_client_encryption_key" "test" {
}
`

func testAccCheckClientEncryptionKeyDestroy(state *terraform.State) error {
	basisTheoryClient := basistheoryClient.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "basistheory_client_encryption_key" {
			continue
		}

		_, err := basisTheoryClient.Keys.Get(context.TODO(), rs.Primary.ID)
		if err == nil {
			return errors.New("Client Encryption Key still exists")
		}
		var notFoundError basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}

func futureDate() string {
	return time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
}

func init() {
	testAccClientEncryptionKeyCreate = fmt.Sprintf(testAccClientEncryptionKeyCreate, futureDate())
}
