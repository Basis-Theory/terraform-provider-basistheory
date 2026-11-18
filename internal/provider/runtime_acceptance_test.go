package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// A Reactor runtime and a Proxy code transform runtime take the same execution_role
// and logs blocks, so each acceptance scenario is written once and run against both.

type runtimeCheck func(address string, runtime string) resource.TestCheckFunc

type runtimeStep struct {
	blocks string
	checks []runtimeCheck
}

func runtimeAttr(attribute string, value string) runtimeCheck {
	return func(address string, runtime string) resource.TestCheckFunc {
		return resource.TestCheckResourceAttr(address, runtime+"."+attribute, value)
	}
}

func runtimeSetElem(attribute string, value string) runtimeCheck {
	return func(address string, runtime string) resource.TestCheckFunc {
		return resource.TestCheckTypeSetElemAttr(address, runtime+"."+attribute, value)
	}
}

func runtimeRule(values map[string]string) runtimeCheck {
	return func(address string, runtime string) resource.TestCheckFunc {
		return resource.TestCheckTypeSetElemNestedAttrs(address, runtime+".execution_role.0.rule.*", values)
	}
}

// runtimeTestSteps turns scenario steps into test steps for one resource. config
// places a step's blocks inside the resource's runtime block, and runtimes are the
// state addresses of every runtime block the checks apply to.
func runtimeTestSteps(steps []runtimeStep, config func(blocks string) string, address string, runtimes ...string) []resource.TestStep {
	var testSteps []resource.TestStep

	for _, step := range steps {
		checks := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(address, "state", "active"),
		}
		for _, runtime := range runtimes {
			for _, check := range step.checks {
				checks = append(checks, check(address, runtime))
			}
		}

		testSteps = append(testSteps, resource.TestStep{
			Config: config(step.blocks),
			Check:  resource.ComposeAggregateTestCheckFunc(checks...),
		})
	}

	return testSteps
}

var executionRoleRulesSteps = []runtimeStep{
	{
		blocks: `
      execution_role {
        rule {
          description = "mask every other token"
          priority    = 2
          container   = "/"
          transform   = "mask"
          permissions = ["token:read"]
        }
      }`,
		checks: []runtimeCheck{
			runtimeAttr("execution_role.0.rule.#", "1"),
			runtimeAttr("execution_role.0.permissions.#", "0"),
			runtimeRule(map[string]string{
				"description":   "mask every other token",
				"priority":      "2",
				"container":     "/",
				"transform":     "mask",
				"permissions.0": "token:read",
			}),
		},
	},
	// Adds a rule and keeps the first one, so the update must send the unchanged rule too.
	{
		blocks: `
      execution_role {
        rule {
          description = "reveal general tokens"
          priority    = 1
          container   = "/general/"
          transform   = "reveal"
          permissions = ["token:read"]
        }
        rule {
          description = "mask every other token"
          priority    = 2
          container   = "/"
          transform   = "mask"
          permissions = ["token:read"]
        }
      }`,
		checks: []runtimeCheck{
			runtimeAttr("execution_role.0.rule.#", "2"),
			runtimeRule(map[string]string{
				"description":   "reveal general tokens",
				"priority":      "1",
				"container":     "/general/",
				"transform":     "reveal",
				"permissions.0": "token:read",
			}),
			runtimeRule(map[string]string{
				"description":   "mask every other token",
				"priority":      "2",
				"container":     "/",
				"transform":     "mask",
				"permissions.0": "token:read",
			}),
		},
	},
}

var executionRolePermissionsSteps = []runtimeStep{
	{
		blocks: `
      execution_role {
        permissions = ["token:create"]
      }`,
		checks: []runtimeCheck{
			runtimeAttr("execution_role.0.permissions.#", "1"),
			runtimeSetElem("execution_role.0.permissions.*", "token:create"),
			runtimeAttr("execution_role.0.rule.#", "0"),
		},
	},
	{
		blocks: `
      execution_role {
        permissions = ["token:create", "token:read"]
      }`,
		checks: []runtimeCheck{
			runtimeAttr("execution_role.0.permissions.#", "2"),
			runtimeSetElem("execution_role.0.permissions.*", "token:create"),
			runtimeSetElem("execution_role.0.permissions.*", "token:read"),
		},
	},
}

var executionRoleSwitchSteps = []runtimeStep{
	{
		blocks: `
      permissions = ["token:create"]`,
		checks: []runtimeCheck{
			runtimeAttr("permissions.#", "1"),
			runtimeAttr("permissions.0", "token:create"),
			runtimeAttr("execution_role.#", "0"),
		},
	},
	{
		blocks: `
      execution_role {
        permissions = ["token:create"]
      }`,
		checks: []runtimeCheck{
			runtimeAttr("permissions.#", "0"),
			runtimeAttr("execution_role.0.permissions.#", "1"),
			runtimeSetElem("execution_role.0.permissions.*", "token:create"),
		},
	},
	{
		blocks: `
      permissions = ["token:create"]`,
		checks: []runtimeCheck{
			runtimeAttr("permissions.#", "1"),
			runtimeAttr("permissions.0", "token:create"),
			runtimeAttr("execution_role.#", "0"),
		},
	},
}

// Each step must also leave an empty plan, which the test framework checks. That
// covers a missing block and logs that are off being equal, in both directions.
var runtimeLogsSteps = []runtimeStep{
	{
		blocks: `
      logs {
        enabled = true
        level   = "debug"
      }`,
		checks: []runtimeCheck{
			runtimeAttr("logs.0.enabled", "true"),
			runtimeAttr("logs.0.level", "debug"),
		},
	},
	{
		blocks: `
      logs {
        enabled = true
        level   = "warn"
      }`,
		checks: []runtimeCheck{
			runtimeAttr("logs.0.enabled", "true"),
			runtimeAttr("logs.0.level", "warn"),
		},
	},
	{
		blocks: `
      logs {
        enabled = false
      }`,
		checks: []runtimeCheck{
			runtimeAttr("logs.0.enabled", "false"),
			runtimeAttr("logs.0.level", "info"),
		},
	},
	{
		blocks: "",
		checks: []runtimeCheck{
			runtimeAttr("logs.#", "0"),
		},
	},
}
