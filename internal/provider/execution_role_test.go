package provider

import (
	"context"
	"encoding/json"
	"testing"

	basistheory "github.com/Basis-Theory/go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testFunctionCode = "module.exports = async function (context) { return context; };"

var testExecutionRoleRule = map[string]interface{}{
	"description": "reveal card numbers",
	"priority":    1,
	"container":   "/pci/",
	"transform":   "reveal",
	"permissions": []interface{}{"token:read"},
}

var testExecutionRoleRuleRequest = map[string]interface{}{
	"description": "reveal card numbers",
	"priority":    float64(1),
	"container":   "/pci/",
	"transform":   "reveal",
	"permissions": []interface{}{"token:read"},
}

// runtimeResource is one place a runtime block can be written. config wraps a runtime
// block in a full configuration, and sent returns the runtime the resource would send.
type runtimeResource struct {
	resource func() *schema.Resource
	config   func(runtime map[string]interface{}) map[string]interface{}
	sent     func(t *testing.T, data *schema.ResourceData) map[string]interface{}
}

var runtimeResources = map[string]runtimeResource{
	"reactor": {
		resource: resourceBasisTheoryReactor,
		config: func(runtime map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"name":    "reactor",
				"code":    testFunctionCode,
				"runtime": []interface{}{runtime},
			}
		},
		sent: func(t *testing.T, data *schema.ResourceData) map[string]interface{} {
			return jsonObject(t, getReactorFromData(data).Runtime)
		},
	},
	"proxy request transform":  proxyRuntimeResource("request_transforms"),
	"proxy response transform": proxyRuntimeResource("response_transforms"),
}

func proxyRuntimeResource(field string) runtimeResource {
	return runtimeResource{
		resource: resourceBasisTheoryProxy,
		config: func(runtime map[string]interface{}) map[string]interface{} {
			return map[string]interface{}{
				"name":            "proxy",
				"destination_url": "https://example.com",
				field: []interface{}{
					map[string]interface{}{
						"type": "code",
						"code": testFunctionCode,
						"options": []interface{}{
							map[string]interface{}{
								"runtime": []interface{}{runtime},
							},
						},
					},
				},
			}
		},
		sent: func(t *testing.T, data *schema.ResourceData) map[string]interface{} {
			proxy := getProxyFromData(data)
			transforms := proxy.RequestTransforms
			if field == "response_transforms" {
				transforms = proxy.ResponseTransforms
			}
			require.Len(t, transforms, 1)

			return jsonObject(t, transforms[0].Options.Runtime)
		},
	}
}

// request returns the runtime sent when the resource is created with a runtime block.
func (r runtimeResource) request(t *testing.T, runtime map[string]interface{}) map[string]interface{} {
	return r.sent(t, schema.TestResourceDataRaw(t, r.resource().Schema, r.config(runtime)))
}

// updateRequest returns the runtime sent when the runtime block changes from before to after.
func (r runtimeResource) updateRequest(t *testing.T, before map[string]interface{}, after map[string]interface{}) map[string]interface{} {
	resource := r.resource()

	created := schema.TestResourceDataRaw(t, resource.Schema, r.config(before))
	created.SetId("id")
	state := created.State()

	diff, err := resource.Diff(context.Background(), state, terraform.NewResourceConfigRaw(r.config(after)), nil)
	require.NoError(t, err)
	data, err := schema.InternalMap(resource.Schema).Data(state, diff)
	require.NoError(t, err)

	return r.sent(t, data)
}

func jsonObject(t *testing.T, value interface{}) map[string]interface{} {
	encoded, err := json.Marshal(value)
	require.NoError(t, err)

	var object map[string]interface{}
	require.NoError(t, json.Unmarshal(encoded, &object))

	return object
}

func executionRoleBlock(role map[string]interface{}) []interface{} {
	return []interface{}{role}
}

func TestExecutionRole_SendsRules(t *testing.T) {
	for name, target := range runtimeResources {
		t.Run(name, func(t *testing.T) {
			runtime := target.request(t, map[string]interface{}{
				"image": "node24",
				"execution_role": executionRoleBlock(map[string]interface{}{
					"rule": []interface{}{testExecutionRoleRule},
				}),
			})

			assert.Equal(t, map[string]interface{}{
				"rules": []interface{}{testExecutionRoleRuleRequest},
			}, runtime["execution_role"])
		})
	}
}

func TestExecutionRole_SendsPermissions(t *testing.T) {
	for name, target := range runtimeResources {
		t.Run(name, func(t *testing.T) {
			runtime := target.request(t, map[string]interface{}{
				"image": "node24",
				"execution_role": executionRoleBlock(map[string]interface{}{
					"permissions": []interface{}{"token:read", "token:use"},
				}),
			})

			role := runtime["execution_role"].(map[string]interface{})
			assert.ElementsMatch(t, []interface{}{"token:read", "token:use"}, role["permissions"])
			assert.NotContains(t, role, "rules")
		})
	}
}

// Whether both lists are allowed together is the API's decision, so both must reach it.
func TestExecutionRole_SendsPermissionsAndRulesTogether(t *testing.T) {
	for name, target := range runtimeResources {
		t.Run(name, func(t *testing.T) {
			runtime := target.request(t, map[string]interface{}{
				"image": "node24",
				"execution_role": executionRoleBlock(map[string]interface{}{
					"permissions": []interface{}{"token:use"},
					"rule":        []interface{}{testExecutionRoleRule},
				}),
			})

			assert.Equal(t, map[string]interface{}{
				"permissions": []interface{}{"token:use"},
				"rules":       []interface{}{testExecutionRoleRuleRequest},
			}, runtime["execution_role"])
		})
	}
}

// runtime.permissions next to an execution role is rejected by the API, not by the provider.
func TestExecutionRole_SendsRuntimePermissionsNextToRole(t *testing.T) {
	for name, target := range runtimeResources {
		t.Run(name, func(t *testing.T) {
			runtime := target.request(t, map[string]interface{}{
				"image":       "node24",
				"permissions": []interface{}{"token:create"},
				"execution_role": executionRoleBlock(map[string]interface{}{
					"permissions": []interface{}{"token:read"},
				}),
			})

			assert.Equal(t, []interface{}{"token:create"}, runtime["permissions"])
			assert.Equal(t, map[string]interface{}{
				"permissions": []interface{}{"token:read"},
			}, runtime["execution_role"])
		})
	}
}

func TestExecutionRole_SendsNoRoleWithoutBlock(t *testing.T) {
	for name, target := range runtimeResources {
		t.Run(name, func(t *testing.T) {
			runtime := target.request(t, map[string]interface{}{
				"image": "node24",
			})

			assert.NotContains(t, runtime, "execution_role")
		})
	}
}

// An update replaces the whole role, so the elements that did not change must be sent too.
func TestExecutionRole_UpdateSendsUnchangedPermissionsAndRules(t *testing.T) {
	secondRule := map[string]interface{}{
		"description": "mask everything else",
		"priority":    2,
		"container":   "/",
		"transform":   "mask",
		"permissions": []interface{}{"token:read"},
	}

	for name, target := range runtimeResources {
		t.Run(name+" permissions", func(t *testing.T) {
			runtime := target.updateRequest(t,
				map[string]interface{}{
					"image": "node24",
					"execution_role": executionRoleBlock(map[string]interface{}{
						"permissions": []interface{}{"token:create"},
					}),
				},
				map[string]interface{}{
					"image": "node24",
					"execution_role": executionRoleBlock(map[string]interface{}{
						"permissions": []interface{}{"token:create", "token:read"},
					}),
				})

			role := runtime["execution_role"].(map[string]interface{})
			assert.ElementsMatch(t, []interface{}{"token:create", "token:read"}, role["permissions"])
		})

		t.Run(name+" rules", func(t *testing.T) {
			runtime := target.updateRequest(t,
				map[string]interface{}{
					"image": "node24",
					"execution_role": executionRoleBlock(map[string]interface{}{
						"rule": []interface{}{testExecutionRoleRule},
					}),
				},
				map[string]interface{}{
					"image": "node24",
					"execution_role": executionRoleBlock(map[string]interface{}{
						"rule": []interface{}{testExecutionRoleRule, secondRule},
					}),
				})

			role := runtime["execution_role"].(map[string]interface{})
			assert.Len(t, role["rules"], 2)
			assert.Contains(t, role["rules"], testExecutionRoleRuleRequest)
		})
	}
}

func TestExecutionRole_ReadsRoleIntoState(t *testing.T) {
	role := &basistheory.ExecutionRole{
		Permissions: []string{"token:use"},
		Rules: []*basistheory.AccessRule{
			{
				Description: getStringPointer("reveal card numbers"),
				Priority:    getIntPointer(1),
				Container:   getStringPointer("/pci/"),
				Transform:   getStringPointer("reveal"),
				Permissions: []string{"token:read"},
			},
		},
	}

	assertRoleInState := func(t *testing.T, data *schema.ResourceData, roleAddress string) {
		assert.Equal(t, []interface{}{"token:use"}, data.Get(roleAddress+".permissions").(*schema.Set).List())

		rules := data.Get(roleAddress + ".rule").(*schema.Set).List()
		require.Len(t, rules, 1)
		rule := rules[0].(map[string]interface{})
		assert.Equal(t, "reveal card numbers", rule["description"])
		assert.Equal(t, 1, rule["priority"])
		assert.Equal(t, "/pci/", rule["container"])
		assert.Equal(t, "reveal", rule["transform"])
		assert.Equal(t, []interface{}{"token:read"}, rule["permissions"].(*schema.Set).List())
	}

	t.Run("reactor", func(t *testing.T) {
		meta := newStubbedMeta(t, func(method string) interface{} {
			return stubbedReactor(map[string]interface{}{
				"image":          "node24",
				"execution_role": jsonObject(t, role),
			})
		})
		data := schema.TestResourceDataRaw(t, resourceBasisTheoryReactor().Schema, map[string]interface{}{})
		data.SetId("reactor-id")

		require.False(t, resourceReactorRead(context.Background(), data, meta).HasError())

		assertRoleInState(t, data, "runtime.0.execution_role.0")
		assert.Empty(t, data.Get("runtime.0.permissions"))
	})

	for _, field := range []string{"request_transforms", "response_transforms"} {
		t.Run("proxy "+field, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, resourceBasisTheoryProxy().Schema, map[string]interface{}{})

			require.NoError(t, data.Set(field, flattenProxyTransforms([]*basistheory.ProxyTransform{
				{
					Type: getStringPointer("code"),
					Code: getStringPointer(testFunctionCode),
					Options: &basistheory.ProxyTransformOptions{
						Runtime: &basistheory.Runtime{
							Image:         getStringPointer("node24"),
							ExecutionRole: role,
						},
					},
				},
			}, nil)))

			assertRoleInState(t, data, field+".0.options.0.runtime.0.execution_role.0")
		})
	}
}
