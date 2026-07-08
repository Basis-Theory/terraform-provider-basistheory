package provider

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	basistheory "github.com/Basis-Theory/go-sdk/v5"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/Basis-Theory/go-sdk/v5/option"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestReactorFinalStateDiagnostics_withoutRequestedError(t *testing.T) {
	actual := reactorFinalStateDiagnostics("rct_123", "failed", &basistheory.Reactor{})

	if len(actual) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(actual))
	}

	expected := "reactor rct_123 reached failed state"
	if actual[0].Summary != expected {
		t.Fatalf("expected %q, got %q", expected, actual[0].Summary)
	}
}

func TestReactorFinalStateDiagnostics_usesActualFinalState(t *testing.T) {
	actual := reactorFinalStateDiagnostics("rct_123", "outdated", &basistheory.Reactor{})

	if len(actual) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(actual))
	}

	expected := "reactor rct_123 reached outdated state"
	if actual[0].Summary != expected {
		t.Fatalf("expected %q, got %q", expected, actual[0].Summary)
	}
}

func TestReactorFinalStateDiagnostics_includesRequestedError(t *testing.T) {
	errorCode := "vulnerabilities_detected"
	errorMessage := "Please update the dependencies listed to resolve security vulnerabilities."
	reactor := &basistheory.Reactor{
		Requested: &basistheory.RequestedReactor{
			ErrorCode:    &errorCode,
			ErrorMessage: &errorMessage,
			ErrorDetails: map[string]interface{}{
				"vulnerabilities": []interface{}{
					map[string]interface{}{
						"name":            "follow-redirects",
						"version":         "1.14.7",
						"severity":        "HIGH",
						"id":              "CVE-2022-0536",
						"dependency_path": []interface{}{"axios", "follow-redirects"},
					},
				},
			},
		},
	}

	actual := reactorFinalStateDiagnostics("rct_123", "outdated", reactor)

	if len(actual) != 1 {
		t.Fatalf("expected one diagnostic, got %d", len(actual))
	}

	expectedParts := []string{
		"reactor rct_123 reached outdated state",
		"Requested Reactor Error Code: vulnerabilities_detected",
		"Requested Reactor Error Message: Please update the dependencies listed to resolve security vulnerabilities.",
		"Requested Reactor Error Details:",
		"\"name\": \"follow-redirects\"",
		"\"version\": \"1.14.7\"",
		"\"severity\": \"HIGH\"",
		"\"id\": \"CVE-2022-0536\"",
		"\"dependency_path\": [",
		"\"axios\"",
		"\"follow-redirects\"",
	}

	for _, expectedPart := range expectedParts {
		if !strings.Contains(actual[0].Summary, expectedPart) {
			t.Fatalf("expected diagnostic to contain %q, got %q", expectedPart, actual[0].Summary)
		}
	}
}

func TestGetReactorFromData_includesRuntimeResolutions(t *testing.T) {
	data := schema.TestResourceDataRaw(t, resourceBasisTheoryReactor().Schema, map[string]interface{}{
		"name": "Terraform reactor with node22 runtime",
		"code": "module.exports = async function (context) { return context; };",
		"runtime": []interface{}{
			map[string]interface{}{
				"image": "node22",
				"dependencies": map[string]interface{}{
					"axios": "1.15.1",
				},
				"resolutions": map[string]interface{}{
					"follow-redirects": "1.15.6",
				},
			},
		},
	})

	reactor := getReactorFromData(data)

	resolutions := reactor.Runtime.Resolutions
	if actual := resolutions["follow-redirects"]; actual == nil || *actual != "1.15.6" {
		t.Fatalf("expected follow-redirects resolution to be 1.15.6, got %v", actual)
	}
}

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

func TestResourceReactorWithNode22Runtime(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckReactorDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccReactorWithNode22Runtime, "terraform_test_reactor_node22"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "name", "Terraform reactor with node22 runtime"),
					resource.TestMatchResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "code", regexp.MustCompile("return")),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "configuration.TEST_FOO", "TEST_FOO"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.image", "node22"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.dependencies.@basis-theory/node-sdk", "4.2.1"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.resolutions.follow-redirects", "1.15.6"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.warm_concurrency", "1"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.timeout", "10"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.resources", "standard"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "state", "active"),
				),
			},
			{
				Config: fmt.Sprintf(testAccReactorWithNode22RuntimeUpdated, "terraform_test_reactor_node22"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "name", "Terraform reactor with node22 runtime"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.image", "node22"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.dependencies.@basis-theory/node-sdk", "4.2.1"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.dependencies.is-odd", "3.0.1"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.resolutions.follow-redirects", "1.15.6"),
					resource.TestCheckResourceAttr(
						"basistheory_reactor.terraform_test_reactor_node22", "runtime.0.resolutions.is-number", "7.0.0"),
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

const testAccReactorWithNode22Runtime = `
resource "basistheory_reactor" "%s" {
  name = "Terraform reactor with node22 runtime"
  code = <<-EOT
            module.exports = async function (context) {
                return {
					"res": {
						  "body": {},
						  "headers": {
							  "X-My-Custom-Header": "will have this value"
						  },
						  "statusCode": 200
					}
				}
            };
        EOT
  configuration = {
    TEST_FOO = "TEST_FOO"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR"
  }
  runtime {
     image = "node22"
	 dependencies = {
		"@basis-theory/node-sdk" = "4.2.1"
	 }
	 resolutions = {
		"follow-redirects" = "1.15.6"
	 }
     warm_concurrency = 1
     timeout = 10
     resources = "standard"
     permissions = ["token:create"] 
  }
}
`

const testAccReactorWithNode22RuntimeUpdated = `
resource "basistheory_reactor" "%s" {
  name = "Terraform reactor with node22 runtime"
  code = <<-EOT
            module.exports = async function (context) {
                return {
					"res": {
						  "body": {},
						  "headers": {
							  "X-My-Custom-Header": "will have this value"
						  },
						  "statusCode": 200,
					}
				}
            };
        EOT
  configuration = {
    TEST_FOO = "TEST_FOO"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR"
  }
  runtime {
     image = "node22"
	 dependencies = {
		"@basis-theory/node-sdk" = "4.2.1"
		"is-odd" = "3.0.1"
	 }
	 resolutions = {
		"follow-redirects" = "1.15.6"
		"is-number" = "7.0.0"
	 }
     warm_concurrency = 1
     timeout = 10
     resources = "standard"
     permissions = ["token:create"] 
  }
}
`

func TestResourceReactor_HandlesGraceful404(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccReactorCreateWithoutApplication, "terraform_test_reactor_404"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("basistheory_reactor.terraform_test_reactor_404", "id"),
					deleteReactorExternally("basistheory_reactor.terraform_test_reactor_404"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func deleteReactorExternally(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client := basistheoryClient.NewClient(
			option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
			option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
		)

		return client.Reactors.Delete(context.TODO(), rs.Primary.ID)
	}
}

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
		if err == nil {
			return fmt.Errorf("reactor %s still exists", rs.Primary.ID)
		}
		var notFoundError *basistheory.NotFoundError
		if !errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}
