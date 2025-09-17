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

func TestResourceProxy(t *testing.T) {
	testAccApplicationName := "terraform_test_application_proxy_test"
	formattedTestAccReactorCreate := fmt.Sprintf(testAccReactorCreateWithoutApplication, "terraform_test_reactor_proxy_test")
	formattedTestAccApplicationCreate := fmt.Sprintf(testAccApplicationCreateWithCreateKeyTrue, testAccApplicationName)
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

func TestResourceProxyWithoutApplication(t *testing.T) {
	formattedTestAccProxyCreate := fmt.Sprintf(testAccProxyCreateWithoutApplication, "post")
	formattedTestAccProxyUpdate := fmt.Sprintf(testAccProxyCreateWithoutApplication, "get")
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: formattedTestAccProxyCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy without Application"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://httpbin.org/post"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "application_id", ""),
				),
			},
			{
				Config: formattedTestAccProxyUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy without Application"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://httpbin.org/get"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "application_id", ""),
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

func TestResourceProxyWithMaskRegexResponseTransform(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(`
	type = "mask"
	matcher = "regex"
	expression = "(.*)"
	replacement = "*"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.type", "mask"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.matcher", "regex"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.expression", "(.*)"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.replacement", "*"),
				),
			},
		},
	})
}

func TestResourceProxyWithMaskChaseStratusPanTransform(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(`
	type = "mask"
	matcher = "chase_stratus_pan"
	replacement = "*"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.type", "mask"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.matcher", "chase_stratus_pan"),
					resource.TestCheckNoResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.expression"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transform.replacement", "*"),
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

func TestResourceProxyResponseTransformInvalidAttribute(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(` 
	keyNotFound = "mask",
	code = "code"`),
				ExpectError: regexp.MustCompile(`invalid transform property of: keyNotFound`),
			},
		},
	})
}

func TestResourceProxyMaskRequiresMatcher(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      buildProxyWithResponseTransformAttributes(`type = "mask"`),
				ExpectError: regexp.MustCompile(`matcher is required when type is mask`),
			},
		},
	})
}

func TestResourceProxyMaskRequiresReplacement(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      buildProxyWithResponseTransformAttributes(`type = "mask"`),
				ExpectError: regexp.MustCompile(`replacement is required when type is mask`),
			},
		},
	})
}

func TestResourceProxyMaskAndRegexRequiresExpression(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(`
	type = "mask"
	matcher = "regex"`),
				ExpectError: regexp.MustCompile(`expression is required when type is mask and matcher is regex`),
			},
		},
	})
}

func TestResourceProxyMaskAndCodeIsNotNull(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(`
	type = "mask"
	code = "invalid"`),
				ExpectError: regexp.MustCompile(`type must be code when code is provided`),
			},
		},
	})
}

func TestResourceProxyCodeAndMatcher(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(`
	type = "code"
	code = "valid"
	matcher = "regex"`),
				ExpectError: regexp.MustCompile(`matcher is not valid when type is code`),
			},
		},
	})
}

func TestResourceProxyCodeAndExpression(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(`
	type = "code"
	code = "valid"
	expression = "(.*)"`),
				ExpectError: regexp.MustCompile(`expression is not valid when type is code`),
			},
		},
	})
}

func TestResourceProxyCodeAndReplacement(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(`
	type = "code"
	code = "valid"
	replacement = "*"`),
				ExpectError: regexp.MustCompile(`replacement is not valid when type is code`),
			},
		},
	})
}

func TestResourceProxyTypeCodeAndCodeIsNil(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithResponseTransformAttributes(`
	type = "code"`),
				ExpectError: regexp.MustCompile(`code is required when type is code`),
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
	type = "code"
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
	type = "code"
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

const testAccProxyCreateWithoutApplication = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name            = "Terraform proxy without Application"
  destination_url = "https://httpbin.org/%s"
  require_auth    = true
  request_transform = {
    code = <<-EOT
              const package = require("abcd");
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
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
	basisTheoryClient := basistheoryClient.NewClient(
		option.WithAPIKey(os.Getenv("BASISTHEORY_API_KEY")),
		option.WithBaseURL(os.Getenv("BASISTHEORY_API_URL")),
	)

	for _, rs := range state.RootModule().Resources {
		if rs.Type != "basistheory_proxy" {
			continue
		}

		_, err := basisTheoryClient.Proxies.Get(context.TODO(), rs.Primary.ID)

		var notFoundError basistheory.NotFoundError
		if errors.As(err, &notFoundError) {
			return err
		}
	}

	return nil
}

func buildProxyWithResponseTransformAttributes(config string) string {
	return fmt.Sprintf(`
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  response_transform = {
    %s
  }
  require_auth = false
}
`, config)
}
