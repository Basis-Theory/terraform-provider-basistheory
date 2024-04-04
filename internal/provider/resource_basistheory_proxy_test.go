package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Basis-Theory/basistheory-go/v5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestResourceProxy(t *testing.T) {
	testAccApplicationName := "terraform_test_application_proxy_test"
	formattedTestAccReactorCreate := fmt.Sprintf(testAccReactorCreateWithoutApplication, "terraform_test_reactor_proxy_test")
	formattedTestAccApplicationCreate := fmt.Sprintf(testAccApplicationCreate, testAccApplicationName)
	formattedTestAccProxyCreate := fmt.Sprintf(testAccProxyCreate, testAccApplicationName)
	formattedTestAccProxyUpdate := fmt.Sprintf(testAccProxyUpdate, testAccApplicationName)
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s\n%s\n%s", formattedTestAccReactorCreate, formattedTestAccApplicationCreate, formattedTestAccProxyCreate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://httpbin.org/post"),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_reactor_id", regexp.MustCompile(testUuidRegex)),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_reactor_id", regexp.MustCompile(testUuidRegex)),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transform.code", regexp.MustCompile("module.exports = async function")),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.code", regexp.MustCompile("module.exports = async function")),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "configuration.TEST_FOO", "TEST_FOO"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR"),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "application_id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "require_auth", "false"),
				),
			},
			{
				Config: fmt.Sprintf("%s\n%s\n%s", formattedTestAccReactorCreate, formattedTestAccApplicationCreate, formattedTestAccProxyUpdate),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy updated name"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://httpbin.org/post"),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_reactor_id", regexp.MustCompile(testUuidRegex)),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_reactor_id", regexp.MustCompile(testUuidRegex)),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transform.code", regexp.MustCompile("const package = require")),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.code", regexp.MustCompile("const package = require")),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "configuration.TEST_FOO", "TEST_FOO_UPDATED"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR_UPDATED"),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "application_id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "require_auth", "true"),
				),
			},
		},
	})
}

func TestResourceProxyWithoutRequireAuth(t *testing.T) {
	formattedTestAccReactorCreate := fmt.Sprintf(testAccReactorCreateWithoutApplication, "terraform_test_reactor_proxy_test")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf("%s\n%s", formattedTestAccReactorCreate, testAccProxyCreateWithoutRequireAuth),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://httpbin.org/post"),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_reactor_id", regexp.MustCompile(testUuidRegex)),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_reactor_id", regexp.MustCompile(testUuidRegex)),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "require_auth", "true"),
				),
			},
		},
	})
}

func TestResourceProxyWithoutReactorIds(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyCreateWithoutReactors,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://httpbin.org/post"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_reactor_id", ""),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_reactor_id", ""),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "require_auth", "true"),
				),
			},
		},
	})
}

func TestResourceProxyInvalidTransformProperty(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccProxyCreateWithInvalidTransformProperty,
				ExpectError: regexp.MustCompile(`invalid transform property of: random`),
			},
		},
	})
}

func TestResourceProxyMissingTransformCode(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccProxyCreateWithMissingTransformCode,
				ExpectError: regexp.MustCompile(`code is required`),
			},
		},
	})
}

func TestResourceProxyEmptyTransformCode(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccProxyCreateWithEmptyTransformCode,
				ExpectError: regexp.MustCompile(`code is required`),
			},
		},
	})
}

const testAccProxyCreate = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_reactor_id = "${basistheory_reactor.terraform_test_reactor_proxy_test.id}"
  response_reactor_id = "${basistheory_reactor.terraform_test_reactor_proxy_test.id}"
  request_transform = {
    code = <<-EOT
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
  response_transform = {
    code = <<-EOT
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
  application_id = "${basistheory_application.%s.id}"
  configuration = {
    TEST_FOO = "TEST_FOO"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR"
  }
  require_auth = false
}
`

const testAccProxyUpdate = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy updated name"
  destination_url = "https://httpbin.org/post"
  request_reactor_id = "${basistheory_reactor.terraform_test_reactor_proxy_test.id}"
  response_reactor_id = "${basistheory_reactor.terraform_test_reactor_proxy_test.id}"
  request_transform = {
    code = <<-EOT
              const package = require("abcd");
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
  response_transform = {
    code = <<-EOT
              const package = require("abcd");
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
  application_id = "${basistheory_application.%s.id}"
  configuration = {
    TEST_FOO = "TEST_FOO_UPDATED"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR_UPDATED"
  }
  require_auth = true
}
`

const testAccProxyCreateWithoutRequireAuth = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_reactor_id = "${basistheory_reactor.terraform_test_reactor_proxy_test.id}"
  response_reactor_id = "${basistheory_reactor.terraform_test_reactor_proxy_test.id}"
}
`

const testAccProxyCreateWithoutReactors = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
}
`

const testAccProxyCreateWithInvalidTransformProperty = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transform = {
    random = "random"
  }
  response_transform = {
    random = "random"
  }
  require_auth = false
}
`

const testAccProxyCreateWithMissingTransformCode = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transform = {
  }
  response_transform = {
  }
  require_auth = false
}
`

const testAccProxyCreateWithEmptyTransformCode = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transform = {
    code = ""
  }
  response_transform = {
    code = ""
  }
  require_auth = false
}
`

func testAccCheckProxyDestroy(state *terraform.State) error {
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
		if rs.Type != "basistheory_proxy" {
			continue
		}

		_, _, err := basisTheoryClient.ProxiesApi.GetById(ctxWithApiKey, rs.Primary.ID).Execute()

		if !strings.Contains(err.Error(), "Not Found") {
			return err
		}
	}

	return nil
}
