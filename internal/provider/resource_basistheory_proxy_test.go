package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	basistheory "github.com/Basis-Theory/go-sdk/v5"
	basistheoryClient "github.com/Basis-Theory/go-sdk/v5/client"
	"github.com/Basis-Theory/go-sdk/v5/option"
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
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.code", regexp.MustCompile("module.exports = async function")),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.type", "code"),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.code", regexp.MustCompile("module.exports = async function")),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.type", "code"),
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
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.code", regexp.MustCompile("const package = require")),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.type", "code"),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.code", regexp.MustCompile("const package = require")),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.type", "code"),
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
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "require_auth", "true"),
				),
			},
		},
	})
}

func TestResourceProxyWithNode22Runtimes(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyCreateWithNode22Runtimes,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy with node 22 runtime"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "configuration.TEST_FOO", "TEST_FOO"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "configuration.TEST_CONFIG_BAR", "TEST_CONFIG_BAR"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "require_auth", "false"),
					// Request transform runtime assertions (now under options)
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.options.0.runtime.0.image", "node22"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.options.0.runtime.0.dependencies.@basis-theory/node-sdk", "4.2.1"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.options.0.runtime.0.warm_concurrency", "1"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.options.0.runtime.0.timeout", "10"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.options.0.runtime.0.resources", "standard"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.options.0.runtime.0.permissions.0", "token:create"),
					// Response transform runtime assertions (now under options)
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.options.0.runtime.0.image", "node22"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.options.0.runtime.0.dependencies.@basis-theory/node-sdk", "4.2.1"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.options.0.runtime.0.warm_concurrency", "1"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.options.0.runtime.0.timeout", "10"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.options.0.runtime.0.resources", "standard"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.options.0.runtime.0.permissions.0", "token:create"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "state", "active"),
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
						"basistheory_proxy.terraform_test_proxy", "require_auth", "true"),
				),
			},
		},
	})
}

func TestResourceProxyWithDisableDetokenization(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccProxyCreateWithDisableDetokenizationTrue,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy with disable_detokenization"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://httpbin.org/post"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "disable_detokenization", "true"),
				),
			},
			{
				Config: testAccProxyCreateWithDisableDetokenizationFalse,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Terraform proxy with disable_detokenization"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://httpbin.org/post"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "disable_detokenization", "false"),
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
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.type", "mask"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.matcher", "regex"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.expression", "(.*)"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.replacement", "*"),
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
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.type", "mask"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.matcher", "chase_stratus_pan"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.expression", ""),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.replacement", "*"),
				),
			},
		},
	})
}

func TestResourceProxyUnsupportedTransformProperty(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccProxyCreateWithUnsupportedTransformProperty,
				ExpectError: regexp.MustCompile(`An argument named "random" is not expected here`),
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
				ExpectError: regexp.MustCompile(`code is required when type is 'code'`),
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
	keyNotFound = "mask"
	code = "code"`),
				ExpectError: regexp.MustCompile(`An argument named "keyNotFound" is not expected here`),
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
				ExpectError: regexp.MustCompile(`matcher is required when type is 'mask'`),
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
				ExpectError: regexp.MustCompile(`replacement is required when type is 'mask'`),
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
				ExpectError: regexp.MustCompile(`(?s)expression is required when type is 'mask'.*matcher is 'regex'`),
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
				ExpectError: regexp.MustCompile(`matcher is required when type is 'mask'`),
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
				ExpectError: regexp.MustCompile(`matcher is not valid when type is 'code'`),
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
				ExpectError: regexp.MustCompile(`expression is not valid when type is 'code'`),
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
				ExpectError: regexp.MustCompile(`replacement is not valid when type is 'code'`),
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
				ExpectError: regexp.MustCompile(`code is required when type is 'code'`),
			},
		},
	})
}

func TestResourceProxyWithTokenizeRequestTransform(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithRequestTransformAttributes(`
	type = "tokenize"
	options {
		identifier = "outputTokenA"
		token = jsonencode({
			type = "card"
			data = "{{ encrypted | json: '$.data' }}"
			metadata = {
				foo = "bar"
				card_holder = "{{ encrypted | json: '$.metadata.card_holder' }}"
			}
		})
	}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.type", "tokenize"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.options.0.identifier", "outputTokenA"),
					// Check that token is a valid JSON string containing the expected structure
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["basistheory_proxy.terraform_test_proxy"]
						if !ok {
							return fmt.Errorf("resource not found: basistheory_proxy.terraform_test_proxy")
						}

						tokenValue, ok := rs.Primary.Attributes["request_transforms.0.options.0.token"]
						if !ok {
							return fmt.Errorf("token attribute not found")
						}

						var token map[string]interface{}
						if err := json.Unmarshal([]byte(tokenValue), &token); err != nil {
							return fmt.Errorf("token is not valid JSON: %v", err)
						}

						if token["type"] != "card" {
							return fmt.Errorf("expected token.type to be 'card', got %v", token["type"])
						}
						if token["data"] != "{{ encrypted | json: '$.data' }}" {
							return fmt.Errorf("expected token.data to be '{{ encrypted | json: '$.data' }}', got %v", token["data"])
						}

						if metadata, ok := token["metadata"].(map[string]interface{}); ok {
							if metadata["foo"] != "bar" {
								return fmt.Errorf("expected metadata.foo to be 'bar', got %v", metadata["foo"])
							}
							if metadata["card_holder"] != "{{ encrypted | json: '$.metadata.card_holder' }}" {
								return fmt.Errorf("expected metadata.card_holder to be '{{ encrypted | json: '$.metadata.card_holder' }}', got %v", metadata["card_holder"])
							}
						} else {
							return fmt.Errorf("expected metadata to be an object, got %v", token["metadata"])
						}

						return nil
					},
				),
			},
		},
	})
}

func TestResourceProxyWithTwoRequestTransforms(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		CheckDestroy:      testAccCheckProxyDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithTwoRequestTransforms(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check first transform (code)
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.type", "code"),
					resource.TestMatchResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.0.code", regexp.MustCompile("module.exports = async function")),
					// Check second transform (tokenize)
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.1.type", "tokenize"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "request_transforms.1.options.0.identifier", "outputTokenB"),
					// Check that token is a valid JSON string containing the expected structure
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["basistheory_proxy.terraform_test_proxy"]
						if !ok {
							return fmt.Errorf("resource not found: basistheory_proxy.terraform_test_proxy")
						}

						tokenValue, ok := rs.Primary.Attributes["request_transforms.1.options.0.token"]
						if !ok {
							return fmt.Errorf("token attribute not found")
						}

						var token map[string]interface{}
						if err := json.Unmarshal([]byte(tokenValue), &token); err != nil {
							return fmt.Errorf("token is not valid JSON: %v", err)
						}

						if token["type"] != "token" {
							return fmt.Errorf("expected token.type to be 'token', got %v", token["type"])
						}
						if token["data"] != "{{ encrypted | json: '$.data' }}" {
							return fmt.Errorf("expected token.data to be '{{ encrypted | json: '$.data' }}', got %v", token["data"])
						}

						return nil
					},
				),
			},
		},
	})
}

func TestResourceProxyWithMultipleResponseTransforms(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildProxyWithMultipleResponseTransforms(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "name", "Test proxy"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "destination_url", "https://echo.basistheory.com/anything"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "require_auth", "false"),
					// Check first response transform (tokenize)
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.type", "tokenize"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.0.options.0.identifier", "cardToken"),
					// Check second response transform (append_json)
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.1.type", "append_json"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.1.options.0.value", "{{ transform_identifier: 'cardToken' | json: '$.id' }}"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.terraform_test_proxy", "response_transforms.1.options.0.location", "$.card_token_id"),
					// Check that token is a valid JSON string containing the expected structure
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["basistheory_proxy.terraform_test_proxy"]
						if !ok {
							return fmt.Errorf("resource not found: basistheory_proxy.terraform_test_proxy")
						}

						tokenValue, ok := rs.Primary.Attributes["response_transforms.0.options.0.token"]
						if !ok {
							return fmt.Errorf("token attribute not found")
						}

						var token map[string]interface{}
						if err := json.Unmarshal([]byte(tokenValue), &token); err != nil {
							return fmt.Errorf("token is not valid JSON: %v", err)
						}

						if token["type"] != "card" {
							return fmt.Errorf("expected token.type to be 'card', got %v", token["type"])
						}
						if token["data"] != "{{ res.json }}" {
							return fmt.Errorf("expected token.data to be '{{ res.json }}', got %v", token["data"])
						}

						if metadata, ok := token["metadata"].(map[string]interface{}); ok {
							if metadata["source"] != "proxy-response-transform" {
								return fmt.Errorf("expected metadata.source to be 'proxy-response-transform', got %v", metadata["source"])
							}
							if metadata["number"] != "123" {
								return fmt.Errorf("expected metadata.number to be '123', got %v", metadata["number"])
							}
						} else {
							return fmt.Errorf("expected Metadata to be an object, got %v", token["Metadata"])
						}

						return nil
					},
				),
			},
		},
	})
}

func TestResourceProxyWithResponseTransformProxyDefinition(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: buildResponseTransformProxyDefinition(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "name", "Response Transform Proxy"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "destination_url", "https://api.bank.com/accounts"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "require_auth", "true"),
					// First transform: tokenize
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "response_transforms.0.type", "tokenize"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "response_transforms.0.options.0.identifier", "responseAccountToken"),
					// Validate token JSON structure
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["basistheory_proxy.response_transform_proxy"]
						if !ok {
							return fmt.Errorf("resource not found: basistheory_proxy.response_transform_proxy")
						}
						tokenValue, ok := rs.Primary.Attributes["response_transforms.0.options.0.token"]
						if !ok {
							return fmt.Errorf("token attribute not found")
						}
						var token map[string]interface{}
						if err := json.Unmarshal([]byte(tokenValue), &token); err != nil {
							return fmt.Errorf("token is not valid JSON: %v", err)
						}
						if token["type"] != "card" {
							return fmt.Errorf("expected token.type to be 'card', got %v", token["type"])
						}
						// Verify nested data placeholders exist
						if data, ok := token["data"].(map[string]interface{}); ok {
							if data["number"] != "{{ res.number }}" {
								return fmt.Errorf("expected data.number to be '{{ res.number }}', got %v", data["number"])
							}
							if data["cvc"] != "{{ res.cvc }}" {
								return fmt.Errorf("expected data.cvc to be '{{ res.cvc }}', got %v", data["cvc"])
							}
							if data["expiration_month"] != "{{ res.expiration_month }}" {
								return fmt.Errorf("expected data.expiration_month to be '{{ res.expiration_month }}', got %v", data["expiration_month"])
							}
							if data["expiration_year"] != "{{ res.expiration_year }}" {
								return fmt.Errorf("expected data.expiration_year to be '{{ res.expiration_year }}', got %v", data["expiration_year"])
							}
						} else {
							return fmt.Errorf("expected data to be an object, got %v", token["data"])
						}
						if metadata, ok := token["metadata"].(map[string]interface{}); ok {
							if metadata["source"] != "proxy-response" {
								return fmt.Errorf("expected metadata.source to be 'proxy-response', got %v", metadata["source"])
							}
							if metadata["property"] != "robert" {
								return fmt.Errorf("expected metadata.property to be 'robert', got %v", metadata["property"])
							}
							if metadata["another"] != "g" {
								return fmt.Errorf("expected metadata.another to be 'g', got %v", metadata["another"])
							}
						} else {
							return fmt.Errorf("expected metadata to be an object, got %v", token["metadata"])
						}
						return nil
					},
					// Second transform: append_json
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "response_transforms.1.type", "append_json"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "response_transforms.1.options.0.value", "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "response_transforms.1.options.0.location", "$.tokenized_account_id"),
					// Third transform: append_header
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "response_transforms.2.type", "append_header"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "response_transforms.2.options.0.value", "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"),
					resource.TestCheckResourceAttr(
						"basistheory_proxy.response_transform_proxy", "response_transforms.2.options.0.location", "X-Account-Token-ID"),
				),
			},
		},
	})
}

func TestResourceProxyMultipleCodeTransforms(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      buildProxyWithMultipleCodeTransforms(),
				ExpectError: regexp.MustCompile(`only one CODE transform is allowed`),
			},
		},
	})
}

func TestResourceProxyDuplicateTokenizeIdentifiers(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      buildProxyWithDuplicateIdentifiers(),
				ExpectError: regexp.MustCompile(`duplicate identifier found.*duplicateId`),
			},
		},
	})
}

func TestResourceProxyTokenizeValidationErrors(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      buildProxyWithTokenizeNoToken(),
				ExpectError: regexp.MustCompile(`token is required in tokenize transform options`),
			},
			{
				Config:      buildProxyWithTokenizeInvalidJSON(),
				ExpectError: regexp.MustCompile(`token must be valid JSON`),
			},
		},
	})
}

func TestResourceProxyAppendTransformValidations(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      buildProxyWithAppendTransformNoValue(),
				ExpectError: regexp.MustCompile(`value is required in append transform options`),
			},
			{
				Config:      buildProxyWithAppendJSONNoLocation(),
				ExpectError: regexp.MustCompile(`location is required for append_json transforms`),
			},
			{
				Config:      buildProxyWithAppendHeaderNoLocation(),
				ExpectError: regexp.MustCompile(`location is required for append_header transforms`),
			},
		},
	})
}

func TestResourceProxyMaskEdgeCases(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { preCheck(t) },
		ProviderFactories: getProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      buildProxyWithMaskChaseStratusAndExpression(),
				ExpectError: regexp.MustCompile(`(?s)expression should not be provided when matcher is.*'chase_stratus_pan'`),
			},
		},
	})
}

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

const testAccProxyCreate = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transforms {
    type = "code"
    code = <<-EOT
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
  response_transforms {
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
  request_transforms {
    type = "code"
    code = <<-EOT
              const package = require("abcd");
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
  response_transforms {
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
}
`

const testAccProxyCreateWithNode22Runtimes = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy with node 22 runtime"
  destination_url = "https://httpbin.org/post"
  request_transforms {
    type = "code"
    code = <<-EOT
              module.exports = async function (context) {
                return context;
              };
          EOT
    options {
      runtime {
        image = "node22"
        dependencies = {
          "@basis-theory/node-sdk" = "4.2.1"
        }
        warm_concurrency = 1
        timeout = 10
        resources = "standard"
        permissions = ["token:create"]
      }
    }
  }
  response_transforms {
    type = "code"
    code = <<-EOT
              module.exports = async function (context) {
                return context;
              };
          EOT
    options {
      runtime {
        image = "node22"
        dependencies = {
          "@basis-theory/node-sdk" = "4.2.1"
        }
        warm_concurrency = 1
        timeout = 10
        resources = "standard"
        permissions = ["token:create"]
      }
    }
  }
  configuration = {
    TEST_FOO = "TEST_FOO"
    TEST_CONFIG_BAR = "TEST_CONFIG_BAR"
  }
  require_auth = false
}
`

const testAccProxyCreateWithoutApplication = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name            = "Terraform proxy without Application"
  destination_url = "https://httpbin.org/%s"
  require_auth    = true
  request_transforms {
    type = "code"
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

const testAccProxyCreateWithDisableDetokenizationTrue = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy with disable_detokenization"
  destination_url = "https://httpbin.org/post"
  disable_detokenization = true
}
`

const testAccProxyCreateWithDisableDetokenizationFalse = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy with disable_detokenization"
  destination_url = "https://httpbin.org/post"
  disable_detokenization = false
}
`

const testAccProxyCreateWithUnsupportedTransformProperty = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transforms {
    random = "random"
  }
  response_transforms {
    random = "random"
  }
  require_auth = false
}
`

const testAccProxyCreateWithMissingTransformCode = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transforms {
	type = "code"
  }
  response_transforms {
	type = "code"
  }
  require_auth = false
}
`

const testAccProxyCreateWithEmptyTransformCode = `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transforms {
    type = "code"
    code = ""
  }
  response_transforms {
    type = "code"
    code = ""
  }
  require_auth = false
}
`

func buildProxyWithResponseTransformAttributes(config string) string {
	return fmt.Sprintf(`
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  response_transforms {
    %s
  }
  require_auth = false
}
`, config)
}

func buildProxyWithRequestTransformAttributes(config string) string {
	return fmt.Sprintf(`
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transforms {
    %s
  }
  require_auth = false
}
`, config)
}

func buildProxyWithTwoRequestTransforms() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Terraform proxy"
  destination_url = "https://httpbin.org/post"
  request_transforms {
    type = "code"
    code = <<-EOT
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
  request_transforms {
    type = "tokenize"
    options {
      identifier = "outputTokenB"
      token = jsonencode({
        type = "token"
        data = "{{ encrypted | json: '$.data' }}"
      })
    }
  }
  require_auth = false
}
`
}

func buildProxyWithMultipleResponseTransforms() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://echo.basistheory.com/anything"
  require_auth = false
  response_transforms {
    type = "tokenize"
    options {
      identifier = "cardToken"
      token = jsonencode({
        type = "card"
        data = "{{ res.json }}"
        metadata = {
          source = "proxy-response-transform"
          number = "123"
        }
      })
    }
  }
  response_transforms {
    type = "append_json"
    options {
      value = "{{ transform_identifier: 'cardToken' | json: '$.id' }}"
      location = "$.card_token_id"
    }
  }
}
`
}

func buildProxyWithMultipleCodeTransforms() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://httpbin.org/post"
  require_auth = false
  request_transforms {
    type = "code"
    code = <<-EOT
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
  request_transforms {
    type = "code"
    code = <<-EOT
              module.exports = async function (context) {
                return context;
              };
          EOT
  }
}
`
}

func buildProxyWithDuplicateIdentifiers() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://httpbin.org/post"
  require_auth = false
  request_transforms {
    type = "tokenize"
    options {
      identifier = "duplicateId"
      token = jsonencode({
        type = "card"
        data = "{{ request | json: '$.card1' }}"
      })
    }
  }
  request_transforms {
    type = "tokenize"
    options  {
      identifier = "duplicateId"
      token = jsonencode({
        type = "token"
        data = "{{ request | json: '$.card2' }}"
      })
    }
  }
}
`
}

func buildProxyWithTokenizeNoToken() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://httpbin.org/post"
  require_auth = false
  request_transforms {
    type = "tokenize"
    options {
      identifier = "testId"
    }
  }
}
`
}

func buildProxyWithTokenizeInvalidJSON() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://httpbin.org/post"
  require_auth = false
  request_transforms {
    type = "tokenize"
    options {
      identifier = "testId"
      token = "invalid-json-string"
    }
  }
}
`
}

func buildProxyWithAppendTransformNoValue() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://httpbin.org/post"
  require_auth = false
  response_transforms {
    type = "append_text"
    options  {
      location = "$.test"
    }
  }
}
`
}

func buildProxyWithAppendJSONNoLocation() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://httpbin.org/post"
  require_auth = false
  response_transforms {
    type = "append_json"
    options  {
      value = "test value"
    }
  }
}
`
}

func buildProxyWithAppendHeaderNoLocation() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://httpbin.org/post"
  require_auth = false
  response_transforms {
    type = "append_header"
    options  {
      value = "test value"
    }
  }
}
`
}

func buildProxyWithMaskChaseStratusAndExpression() string {
	return `
resource "basistheory_proxy" "terraform_test_proxy" {
  name = "Test proxy"
  destination_url = "https://httpbin.org/post"
  require_auth = false
  response_transforms {
    type = "mask"
    matcher = "chase_stratus_pan"
    expression = "(.*)"
    replacement = "*"
  }
}
`
}


func buildResponseTransformProxyDefinition() string {
	return `
resource "basistheory_proxy" "response_transform_proxy" {
  name            = "Response Transform Proxy"
  destination_url = "https://api.bank.com/accounts"
  require_auth    = true

  # Response transforms - executed in order on outgoing responses
  response_transforms {
    # Tokenize sensitive account data from response
    type = "tokenize"
    options {
      identifier = "responseAccountToken"
      token = jsonencode({
        type = "card"
        data = {
          "number" : "{{ res.number }}",
          "cvc" : "{{ res.cvc }}",
          "expiration_month" : "{{ res.expiration_month }}",
          "expiration_year" : "{{ res.expiration_year }}"
        },
        metadata = {
          source = "proxy-response"
          property = "robert"
          another = "g"
        }
      })
    }
  }

  response_transforms {
    # Replace account number with token ID in response JSON
    type = "append_json"
    options {
      value    = "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"
      location = "$.tokenized_account_id"
    }
  }

  response_transforms {
    # Add response header with token reference
    type = "append_header"
    options {
      value    = "{{ transform_identifier: 'responseAccountToken' | json: '$.id' }}"
      location = "X-Account-Token-ID"
    }
  }
}
`
}
