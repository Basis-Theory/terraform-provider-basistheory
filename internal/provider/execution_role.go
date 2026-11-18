package provider

import (
	basistheory "github.com/Basis-Theory/go-sdk/v8"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var accessRuleTransforms = []string{"mask", "redact", "reveal"}

// accessRuleResource is the only schema for an access rule. Application rules and
// execution role rules both use it, so the two cannot drift apart.
func accessRuleResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"description": {
				Description: "A description of this Access Rule",
				Type:        schema.TypeString,
				Required:    true,
			},
			"priority": {
				Description:  "The priority of this Access Rule, beginning with 1 and higher values having lower precedence",
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"container": {
				Description: "The container of Tokens this rule is scoped to",
				Type:        schema.TypeString,
				Required:    true,
			},
			"transform": {
				Description:  "The transform to apply to accessed Tokens",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(accessRuleTransforms, false),
			},
			"permissions": {
				Description: "List of permissions to grant on this Access Rule",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func executionRoleSchema() *schema.Schema {
	return &schema.Schema{
		Description: "Execution role: the access the code runs with, set through `permissions` or `rule` blocks. " +
			"Declare it here before, or together with, setting it through the API, Portal or CLI: " +
			"when the `runtime` block is set and this block is omitted, an execution role set elsewhere is removed on the next apply that updates the resource. " +
			"Provider versions that do not have this block remove it the same way, so upgrade before setting one",
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"permissions": {
					Description: "Permissions granted to the execution role",
					Type:        schema.TypeSet,
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"rule": {
					Description: "Access rules granted to the execution role",
					Type:        schema.TypeSet,
					Optional:    true,
					MaxItems:    20,
					Elem:        accessRuleResource(),
				},
			},
		},
	}
}

func expandStringSet(raw interface{}) []string {
	set, ok := raw.(*schema.Set)
	if !ok {
		return nil
	}

	var values []string
	for _, value := range set.List() {
		values = append(values, value.(string))
	}

	return values
}

func expandAccessRules(raw interface{}) []*basistheory.AccessRule {
	set, ok := raw.(*schema.Set)
	if !ok {
		return nil
	}

	var rules []*basistheory.AccessRule
	for _, dataRule := range set.List() {
		ruleMap := dataRule.(map[string]interface{})

		rules = append(rules, &basistheory.AccessRule{
			Description: getStringPointer(ruleMap["description"]),
			Priority:    getIntPointer(ruleMap["priority"]),
			Container:   getStringPointer(ruleMap["container"]),
			Transform:   getStringPointer(ruleMap["transform"]),
			Permissions: expandStringSet(ruleMap["permissions"]),
		})
	}

	return rules
}

func flattenAccessRuleData(accessRules []*basistheory.AccessRule) []interface{} {
	if accessRules != nil {
		var flattenedAccessRules []interface{}

		for _, rule := range accessRules {
			flattenedAccessRule := make(map[string]interface{})

			flattenedAccessRule["description"] = rule.Description
			flattenedAccessRule["priority"] = rule.Priority
			flattenedAccessRule["container"] = rule.Container
			flattenedAccessRule["transform"] = rule.Transform
			flattenedAccessRule["permissions"] = rule.Permissions

			flattenedAccessRules = append(flattenedAccessRules, flattenedAccessRule)
		}

		return flattenedAccessRules
	}

	return make([]interface{}, 0)
}

// expandExecutionRole copies permissions and rules independently. Which combinations
// are valid is the API's decision, so a role with both is sent with both.
func expandExecutionRole(raw interface{}) *basistheory.ExecutionRole {
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}

	role := &basistheory.ExecutionRole{}
	// An empty block arrives as a nil element. It is still sent, so the API rejects it.
	if roleMap, ok := list[0].(map[string]interface{}); ok {
		role.Permissions = expandStringSet(roleMap["permissions"])
		role.Rules = expandAccessRules(roleMap["rule"])
	}

	return role
}

func flattenExecutionRole(role *basistheory.ExecutionRole) []interface{} {
	if role == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"permissions": role.Permissions,
			"rule":        flattenAccessRuleData(role.Rules),
		},
	}
}
