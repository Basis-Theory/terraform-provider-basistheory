package provider

import (
	"fmt"
	basistheory "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/Basis-Theory/go-sdk/v5/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/joho/godotenv"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getAccProvider() *schema.Provider {
	missingEnvVars := getMissingEnvVars()

	if missingEnvVars != nil {
		fmt.Printf("%v must be set before running acceptance tests.", strings.Join(missingEnvVars, ", "))
		os.Exit(1)
	}

	userAgent := fmt.Sprintf("HashiCorp Terraform/%s (+https://www.terraform.io) Terraform Plugin SDK/%s", schema.Provider{}.TerraformVersion, meta.SDKVersionString())

	return BasisTheoryProvider(newTestClient(userAgent))()
}

func newTestClient(userAgent string) *basistheory.Client {
	return basistheory.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
		option.WithHTTPHeader(map[string][]string{
			"User-Agent": {userAgent},
		}),
		option.WithHTTPClient(
			&http.Client{
				Timeout: 5 * time.Second,
			},
		),
	)
}

func getProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"basistheory": func() (*schema.Provider, error) {
			return getAccProvider(), nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := getAccProvider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func preCheck(t *testing.T) {
	missingEnvVars := getMissingEnvVars()

	if missingEnvVars != nil {
		t.Fatalf("%v must be set before running acceptance tests.", missingEnvVars)
	}
}

func getMissingEnvVars() []string {
	var requiredEnvironmentVariables = []string{
		"BASISTHEORY_API_KEY",
		"BASISTHEORY_API_URL",
	}
	var missingEnvVars []string
	_ = godotenv.Load("../../.env.local")

	for _, requiredEnvironmentVariable := range requiredEnvironmentVariables {
		if value := os.Getenv(requiredEnvironmentVariable); value == "" {
			missingEnvVars = append(missingEnvVars, requiredEnvironmentVariable)
		}
	}

	return missingEnvVars
}
