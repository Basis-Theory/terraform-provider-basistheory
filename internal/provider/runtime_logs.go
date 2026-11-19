package provider

import (
	basistheory "github.com/Basis-Theory/go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const runtimeLogsDefaultLevel = "info"

var runtimeLogLevels = []string{"debug", "info", "warn", "error"}

func runtimeLogsSchema() *schema.Schema {
	return &schema.Schema{
		Description: "Runtime logs. When the `runtime` block is set and this block is omitted, logs are off at level `info`: " +
			"logs turned on outside Terraform, or a level other than `info` set there, are reset on the next apply that updates the resource. " +
			"To keep them, add this block before applying",
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"enabled": {
					Description: "Whether runtime logs are captured. Defaults to false, so a block that only sets `level` captures nothing",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
				"level": {
					Description:  "Level captured, together with every more severe level: `debug`, `info`, `warn` or `error`. Defaults to `info`",
					Type:         schema.TypeString,
					Optional:     true,
					Default:      runtimeLogsDefaultLevel,
					ValidateFunc: validation.StringInSlice(runtimeLogLevels, false),
				},
			},
		},
	}
}

func expandRuntimeLogs(raw interface{}) *basistheory.RuntimeLogOptions {
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}

	// An empty block arrives as a nil element and means the defaults.
	logsMap, _ := list[0].(map[string]interface{})
	enabled, _ := logsMap["enabled"].(bool)
	level, _ := logsMap["level"].(string)
	if level == "" {
		level = runtimeLogsDefaultLevel
	}

	return &basistheory.RuntimeLogOptions{
		Enabled: &enabled,
		Level:   &level,
	}
}

// flattenRuntimeLogs keeps a missing block and logs that are off equal. The API returns
// logs for every function, off at the default level when they were never set, so that
// default is only written to state when the block is already in the resource data.
func flattenRuntimeLogs(logs *basistheory.RuntimeLogOptions, blockInData bool) []interface{} {
	if logs == nil {
		return nil
	}

	enabled := logs.Enabled != nil && *logs.Enabled
	level := runtimeLogsDefaultLevel
	if logs.Level != nil && *logs.Level != "" {
		level = *logs.Level
	}

	if !blockInData && !enabled && level == runtimeLogsDefaultLevel {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"enabled": enabled,
			"level":   level,
		},
	}
}

func hasBlock(data *schema.ResourceData, path string) bool {
	block, ok := data.Get(path).([]interface{})

	return ok && len(block) > 0
}
