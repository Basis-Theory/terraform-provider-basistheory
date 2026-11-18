package provider

import (
	"encoding/json"
	"fmt"
	basistheory "github.com/Basis-Theory/go-sdk/v8/client"
	"github.com/Basis-Theory/go-sdk/v8/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/meta"
	"github.com/joho/godotenv"
	"net/http"
	"net/http/httptest"
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

// newStubbedMeta returns provider meta whose client talks to a local server. The
// server answers every request with respond's result as JSON, so a test can drive a
// resource function without the API.
func newStubbedMeta(t *testing.T, respond func(method string) interface{}) map[string]interface{} {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(respond(request.Method))
	}))
	t.Cleanup(server.Close)

	return map[string]interface{}{
		"client": basistheory.NewClient(
			option.WithAPIKey("stubbed"),
			option.WithBaseURL(server.URL),
		),
	}
}

func stubbedReactor(runtime map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":         "reactor-id",
		"name":       "reactor",
		"code":       testFunctionCode,
		"state":      "active",
		"created_at": "2026-01-01T00:00:00Z",
		"runtime":    runtime,
	}
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
