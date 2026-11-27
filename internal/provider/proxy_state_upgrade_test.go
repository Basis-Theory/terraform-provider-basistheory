package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

func TestProxyStateUpgradePreservesTransformOptions(t *testing.T) {
	for _, field := range []string{"request_transforms", "response_transforms"} {
		t.Run(field, func(t *testing.T) {
			provider := BasisTheoryProvider(nil)()
			server := schema.NewGRPCProviderServer(provider)
			const resourceName = "basistheory_proxy"

			// In v3, an explicit empty options map was valid on a code transform.
			state := fmt.Sprintf(`{
				"id": "legacy-proxy",
				"name": "legacy-proxy",
				"destination_url": "https://example.com",
				"%s": [
					{"type": "code", "code": "return response;", "options": {}},
					{"type": "code", "code": "return response;"},
					{"type": "append_header", "options": {"location": "X-Example", "value": "sample"}}
				]
			}`, field)
			result, err := server.UpgradeResourceState(context.Background(), &tfprotov5.UpgradeResourceStateRequest{
				TypeName: resourceName,
				Version:  0,
				RawState: &tfprotov5.RawState{JSON: []byte(state)},
			})
			require.NoError(t, err)
			if len(result.Diagnostics) > 0 {
				t.Fatalf("state upgrade failed: %s %s", result.Diagnostics[0].Summary, result.Diagnostics[0].Detail)
			}
			require.NotNil(t, result.UpgradedState)

			stateType := provider.ResourcesMap[resourceName].CoreConfigSchema().ImpliedType()
			upgraded, err := msgpack.Unmarshal(result.UpgradedState.MsgPack, stateType)
			require.NoError(t, err)
			require.Equal(t, "legacy-proxy", upgraded.GetAttr("id").AsString())
			transforms := upgraded.GetAttr(field).AsValueSlice()
			require.Len(t, transforms, 3)
			require.Equal(t, "code", transforms[0].GetAttr("type").AsString())
			require.Zero(t, transforms[0].GetAttr("options").LengthInt())
			require.Zero(t, transforms[1].GetAttr("options").LengthInt())
			populated := transforms[2].GetAttr("options").Index(cty.NumberIntVal(0))
			require.Equal(t, "X-Example", populated.GetAttr("location").AsString())
			require.Equal(t, "sample", populated.GetAttr("value").AsString())
		})
	}
}
