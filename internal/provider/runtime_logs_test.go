package provider

import (
	"context"
	"testing"

	basistheory "github.com/Basis-Theory/go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func logsBlock(logs map[string]interface{}) []interface{} {
	return []interface{}{logs}
}

func runtimeLogs(enabled bool, level string) *basistheory.RuntimeLogOptions {
	return &basistheory.RuntimeLogOptions{Enabled: &enabled, Level: &level}
}

func TestRuntimeLogs_SendsEnabledAndLevel(t *testing.T) {
	for name, target := range runtimeResources {
		t.Run(name, func(t *testing.T) {
			runtime := target.request(t, map[string]interface{}{
				"image": "node24",
				"logs":  logsBlock(map[string]interface{}{"enabled": true, "level": "debug"}),
			})

			assert.Equal(t, map[string]interface{}{"enabled": true, "level": "debug"}, runtime["logs"])
		})
	}
}

func TestRuntimeLogs_SendsLogsOffWhenOnlyLevelIsSet(t *testing.T) {
	for name, target := range runtimeResources {
		t.Run(name, func(t *testing.T) {
			runtime := target.request(t, map[string]interface{}{
				"image": "node24",
				"logs":  logsBlock(map[string]interface{}{"level": "warn"}),
			})

			assert.Equal(t, map[string]interface{}{"enabled": false, "level": "warn"}, runtime["logs"])
		})
	}
}

func TestRuntimeLogs_SendsNoLogsWithoutBlock(t *testing.T) {
	for name, target := range runtimeResources {
		t.Run(name, func(t *testing.T) {
			runtime := target.request(t, map[string]interface{}{
				"image": "node24",
			})

			assert.NotContains(t, runtime, "logs")
		})
	}
}

func TestRuntimeLogs_FlattenTreatsOffAndMissingAsEqual(t *testing.T) {
	block := func(enabled bool, level string) []interface{} {
		return []interface{}{map[string]interface{}{"enabled": enabled, "level": level}}
	}

	for name, test := range map[string]struct {
		logs        *basistheory.RuntimeLogOptions
		blockInData bool
		expected    []interface{}
	}{
		"off at the default level without a block":    {runtimeLogs(false, "info"), false, nil},
		"off at the default level with a block":       {runtimeLogs(false, "info"), true, block(false, "info")},
		"on without a block":                          {runtimeLogs(true, "info"), false, block(true, "info")},
		"off at another level without a block":        {runtimeLogs(false, "debug"), false, block(false, "debug")},
		"on with a block":                             {runtimeLogs(true, "debug"), true, block(true, "debug")},
		"off with no level returned, without a block": {&basistheory.RuntimeLogOptions{}, false, nil},
		"no logs returned":                            {nil, true, nil},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, test.expected, flattenRuntimeLogs(test.logs, test.blockInData))
		})
	}
}

func TestRuntimeLogs_ReactorRead(t *testing.T) {
	for name, test := range map[string]struct {
		configuredLogs []interface{}
		returnedLogs   map[string]interface{}
		expectedState  []interface{}
	}{
		"off without a block stays out of state": {
			returnedLogs:  map[string]interface{}{"enabled": false, "level": "info"},
			expectedState: []interface{}{},
		},
		"off keeps an explicit block": {
			configuredLogs: logsBlock(map[string]interface{}{"enabled": false}),
			returnedLogs:   map[string]interface{}{"enabled": false, "level": "info"},
			expectedState:  logsBlock(map[string]interface{}{"enabled": false, "level": "info"}),
		},
		"turned on outside Terraform shows in state": {
			returnedLogs:  map[string]interface{}{"enabled": true, "level": "debug"},
			expectedState: logsBlock(map[string]interface{}{"enabled": true, "level": "debug"}),
		},
	} {
		t.Run(name, func(t *testing.T) {
			meta := newStubbedMeta(t, func(method string) interface{} {
				return stubbedReactor(map[string]interface{}{
					"image": "node24",
					"logs":  test.returnedLogs,
				})
			})
			runtime := map[string]interface{}{"image": "node24"}
			if test.configuredLogs != nil {
				runtime["logs"] = test.configuredLogs
			}
			data := schema.TestResourceDataRaw(t, resourceBasisTheoryReactor().Schema, map[string]interface{}{
				"name":    "reactor",
				"code":    testFunctionCode,
				"runtime": []interface{}{runtime},
			})
			data.SetId("reactor-id")

			require.False(t, resourceReactorRead(context.Background(), data, meta).HasError())

			assert.Equal(t, test.expectedState, data.Get("runtime.0.logs"))
		})
	}
}

func TestRuntimeLogs_ProxyStateFollowsEachTransformsBlock(t *testing.T) {
	for _, field := range []string{"request_transforms", "response_transforms"} {
		t.Run(field, func(t *testing.T) {
			transformConfig := func(runtime map[string]interface{}) map[string]interface{} {
				return map[string]interface{}{
					"type": "code",
					"code": testFunctionCode,
					"options": []interface{}{
						map[string]interface{}{"runtime": []interface{}{runtime}},
					},
				}
			}
			data := schema.TestResourceDataRaw(t, resourceBasisTheoryProxy().Schema, map[string]interface{}{
				"name":            "proxy",
				"destination_url": "https://example.com",
				field: []interface{}{
					transformConfig(map[string]interface{}{"image": "node24"}),
					transformConfig(map[string]interface{}{
						"image": "node24",
						"logs":  logsBlock(map[string]interface{}{"enabled": false}),
					}),
				},
			})
			returnedTransform := &basistheory.ProxyTransform{
				Type: getStringPointer("code"),
				Code: getStringPointer(testFunctionCode),
				Options: &basistheory.ProxyTransformOptions{
					Runtime: &basistheory.Runtime{
						Image: getStringPointer("node24"),
						Logs:  runtimeLogs(false, "info"),
					},
				},
			}
			proxy := &basistheory.Proxy{ID: getStringPointer("proxy-id")}
			if field == "request_transforms" {
				proxy.RequestTransforms = []*basistheory.ProxyTransform{returnedTransform, returnedTransform}
			} else {
				proxy.ResponseTransforms = []*basistheory.ProxyTransform{returnedTransform, returnedTransform}
			}

			require.False(t, setProxyState(data, proxy).HasError())

			assert.Empty(t, data.Get(field+".0.options.0.runtime.0.logs"))
			assert.Equal(t,
				logsBlock(map[string]interface{}{"enabled": false, "level": "info"}),
				data.Get(field+".1.options.0.runtime.0.logs"))
		})
	}
}

// The API returns logs for every code transform that has options, also for one with no
// runtime configured. Those default logs must not put a runtime block into state.
func TestRuntimeLogs_ProxyDefaultLogsAloneAddNoRuntimeBlock(t *testing.T) {
	flattened := flattenProxyTransforms([]*basistheory.ProxyTransform{
		{
			Type: getStringPointer("code"),
			Code: getStringPointer(testFunctionCode),
			Options: &basistheory.ProxyTransformOptions{
				Runtime: &basistheory.Runtime{Logs: runtimeLogs(false, "info")},
			},
		},
	}, nil)

	assert.NotContains(t, flattened[0], "options")
}

func TestRuntimeLogs_ProxyLogsOnlyRoundTrip(t *testing.T) {
	for _, field := range []string{"request_transforms", "response_transforms"} {
		t.Run(field, func(t *testing.T) {
			for name, logs := range map[string]struct {
				enabled bool
				level   string
			}{
				"enabled":          {true, "debug"},
				"disabled default": {false, "info"},
			} {
				t.Run(name, func(t *testing.T) {
					target := proxyRuntimeResource(field)
					expected := map[string]interface{}{"enabled": logs.enabled, "level": logs.level}
					data := schema.TestResourceDataRaw(t, target.resource().Schema, target.config(map[string]interface{}{
						"logs": logsBlock(expected),
					}))
					require.Equal(t, expected, target.sent(t, data)["logs"])

					transform := &basistheory.ProxyTransform{
						Type: getStringPointer("code"),
						Code: getStringPointer(testFunctionCode),
						Options: &basistheory.ProxyTransformOptions{
							Runtime: &basistheory.Runtime{Logs: runtimeLogs(logs.enabled, logs.level)},
						},
					}
					proxy := &basistheory.Proxy{ID: getStringPointer("proxy-id")}
					if field == "request_transforms" {
						proxy.RequestTransforms = []*basistheory.ProxyTransform{transform}
					} else {
						proxy.ResponseTransforms = []*basistheory.ProxyTransform{transform}
					}

					require.False(t, setProxyState(data, proxy).HasError())
					after := getProxyFromData(data)
					transforms := after.RequestTransforms
					if field == "response_transforms" {
						transforms = after.ResponseTransforms
					}
					require.Len(t, transforms, 1)
					require.NotNil(t, transforms[0].Options)
					require.NotNil(t, transforms[0].Options.Runtime)
					assert.Equal(t, runtimeLogs(logs.enabled, logs.level), transforms[0].Options.Runtime.Logs)
				})
			}
		})
	}
}

func TestRuntimeLogs_LevelAcceptsOnlyKnownLevels(t *testing.T) {
	validate := runtimeLogsSchema().Elem.(*schema.Resource).Schema["level"].ValidateFunc

	for _, level := range []string{"debug", "info", "warn", "error"} {
		_, errs := validate(level, "level")
		assert.Empty(t, errs, level)
	}

	_, errs := validate("verbose", "level")
	assert.NotEmpty(t, errs)
}
