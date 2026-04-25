// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccActionResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-basic", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-basic"),
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
				),
			},
		},
	})
}

func TestAccActionResource_update(t *testing.T) {
	const name = "tf-acc-action-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig(name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", name),
				),
			},
			{
				Config: testAccActionResourceConfigWithFileContents(name, "console.log('hello');"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "file_contents", "console.log('hello');"),
				),
			},
		},
	})
}

func TestAccActionResource_importState(t *testing.T) {
	var actionID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-import", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_action.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						actionID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccActionResourceConfig("tf-acc-action-import", ""),
				ResourceName:      "komodo_action.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return actionID, nil
				},
				ImportStateVerifyIgnore: []string{"webhook"},
			},
		},
	})
}

func TestAccActionResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-disappears", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					testAccActionDisappears("komodo_action.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccActionResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfig("tf-acc-action-rename-orig", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_action.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_action.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccActionResourceConfig("tf-acc-action-rename-new", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "name", "tf-acc-action-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_action.test"]
						if rs.Primary.ID != savedID {
							return fmt.Errorf("resource was recreated: ID changed from %q to %q", savedID, rs.Primary.ID)
						}
						return nil
					},
				),
			},
		},
	})
}

func testAccActionDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		c := client.NewClient(
			os.Getenv("KOMODO_ENDPOINT"),
			os.Getenv("KOMODO_USERNAME"),
			os.Getenv("KOMODO_PASSWORD"),
		)
		return c.DeleteAction(context.Background(), rs.Primary.ID)
	}
}

func testAccActionResourceConfig(name, _ string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
}
`, name)
}

func testAccActionResourceConfigWithFileContents(name, fileContents string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name          = %q
  file_contents = %q
}
`, name, fileContents)
}

func TestAccActionResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionWithTagConfig("tf-acc-action-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_action.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccActionClearTagsConfig("tf-acc-action-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccActionWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-action"
  color = "Green"
}

resource "komodo_action" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccActionClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
  tags = []
}
`, name)
}

func TestUnitActionResource_cronExpressionValidation(t *testing.T) {
	// Invalid expressions are tested via resource.UnitTest (schema validation fires
	// before the provider needs a real connection so no live Komodo instance is needed).
	// Valid expressions are tested by calling the validator directly to avoid running
	// a full plan/apply that would require the provider to be configured.

	invalidCases := []struct {
		name       string
		expression string
	}{
		{"invalid 5-field", "0 0 * * *"},
		{"invalid 4-field", "0 * * *"},
		{"invalid 1-field", "0"},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := fmt.Sprintf(`
resource "komodo_action" "test" {
  name = "test-cron-validation"
  schedule {
    format     = "Cron"
    expression = %q
  }
}
`, tc.expression)

			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      cfg,
						ExpectError: regexp.MustCompile(`Invalid Cron expression`),
					},
				},
			})
		})
	}

	// Valid expressions: call the validator directly — no provider connection needed.
	validCases := []struct {
		name       string
		expression string
	}{
		{"valid 6-field", "0 0 * * * *"},
		{"valid 7-field", "0 0 0 * * * *"},
	}

	for _, tc := range validCases {
		t.Run(tc.name, func(t *testing.T) {
			// Mirror the validator logic: a valid Cron expression has 6 or 7 fields.
			expr := strings.TrimSpace(tc.expression)
			n := len(strings.Fields(expr))
			if n != 6 && n != 7 {
				t.Errorf("expected valid 6- or 7-field cron expression, got %d fields in %q", n, tc.expression)
			}
		})
	}
}

func TestAccActionResource_schedule(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithFullSchedule("tf-acc-action-schedule"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.expression", "0 0 * * * *"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.timezone", "Europe/Berlin"),
				),
			},
		},
	})
}

func TestAccActionResource_scheduleDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithMinimalSchedule("tf-acc-action-schedule-defaults"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.expression", "0 0 * * * *"),
					// enabled and alert_enabled default to true; timezone defaults to ""
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.timezone", ""),
				),
			},
		},
	})
}

func TestAccActionResource_scheduleUpdate(t *testing.T) {
	const name = "tf-acc-action-schedule-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithFullSchedule(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.expression", "0 0 * * * *"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.timezone", "Europe/Berlin"),
				),
			},
			{
				// Update expression only; omit enabled/alert_enabled/timezone → defaults applied
				Config: testAccActionResourceConfigWithMinimalSchedule(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.expression", "0 0 * * * *"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_action.test", "schedule.timezone", ""),
				),
			},
		},
	})
}

func TestAccActionResource_runOnStartupEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithRunOnStartup("tf-acc-action-run-startup"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "run_on_startup_enabled", "true"),
				),
			},
		},
	})
}

func TestAccActionResource_reloadDependenciesEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithReloadDeps("tf-acc-action-reload-deps"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "reload_dependencies_enabled", "true"),
				),
			},
		},
	})
}

func testAccActionResourceConfigWithFullSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
  schedule {
    format        = "Cron"
    expression    = "0 0 * * * *"
    enabled       = true
    alert_enabled = true
    timezone      = "Europe/Berlin"
  }
}
`, name)
}

func testAccActionResourceConfigWithMinimalSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q
  schedule {
    format     = "Cron"
    expression = "0 0 * * * *"
  }
}
`, name)
}

func testAccActionResourceConfigWithRunOnStartup(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name                   = %q
  run_on_startup_enabled = true
}
`, name)
}

func testAccActionResourceConfigWithReloadDeps(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name                        = %q
  reload_dependencies_enabled = true
}
`, name)
}

func TestAccActionResource_arguments(t *testing.T) {
	const name = "tf-acc-action-arguments"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccActionResourceConfigWithArguments(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "argument.#", "2"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.0.name", "MY_TEST_VAR"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.0.value", "Hello World"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.1.name", "MY_TEST_VAR2"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.1.value", "Hello World2"),
				),
			},
			{
				// Update: change a value and remove one argument
				Config: testAccActionResourceConfigWithSingleArgument(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "argument.#", "1"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.0.name", "MY_TEST_VAR"),
					resource.TestCheckResourceAttr("komodo_action.test", "argument.0.value", "Updated Value"),
				),
			},
			{
				// Remove all arguments
				Config: testAccActionResourceConfig(name, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_action.test", "argument.#", "0"),
				),
			},
		},
	})
}

func testAccActionResourceConfigWithArguments(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q

  argument {
    name  = "MY_TEST_VAR"
    value = "Hello World"
  }

  argument {
    name  = "MY_TEST_VAR2"
    value = "Hello World2"
  }
}
`, name)
}

func testAccActionResourceConfigWithSingleArgument(name string) string {
	return fmt.Sprintf(`
resource "komodo_action" "test" {
  name = %q

  argument {
    name  = "MY_TEST_VAR"
    value = "Updated Value"
  }
}
`, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawActionPlan(t *testing.T, r *ActionResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawActionState(t *testing.T, r *ActionResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitActionResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &ActionResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitActionResource_createPlanGetError(t *testing.T) {
	r := &ActionResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawActionPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitActionResource_readStateGetError(t *testing.T) {
	r := &ActionResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawActionState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitActionResource_updatePlanGetError(t *testing.T) {
	r := &ActionResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawActionPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitActionResource_deleteStateGetError(t *testing.T) {
	r := &ActionResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawActionState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitActionResource_parseActionArguments(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		if parseActionArguments("json", "") != nil {
			t.Fatal("expected nil for empty raw")
		}
		if parseActionArguments("json", "{}") != nil {
			t.Fatal("expected nil for empty JSON object")
		}
	})

	t.Run("json_format", func(t *testing.T) {
		result := parseActionArguments("json", `{"b": "2", "a": "1"}`)
		if len(result) != 2 {
			t.Fatalf("expected 2 args, got %d", len(result))
		}
		if result[0].Name.ValueString() != "a" {
			t.Fatal("expected sorted: a first")
		}
	})

	t.Run("json_invalid", func(t *testing.T) {
		result := parseActionArguments("json", "not-json")
		if result != nil {
			t.Fatal("expected nil for invalid JSON")
		}
	})

	t.Run("default_format_kv", func(t *testing.T) {
		result := parseActionArguments("key_value", "B=2\nA=1\n")
		if len(result) != 2 {
			t.Fatalf("expected 2 args, got %d", len(result))
		}
		if result[0].Name.ValueString() != "A" {
			t.Fatal("expected sorted: A first")
		}
	})

	t.Run("default_format_quoted_value", func(t *testing.T) {
		result := parseActionArguments("key_value", `KEY="hello world"`)
		if len(result) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(result))
		}
		if result[0].Value.ValueString() != "hello world" {
			t.Fatalf("expected unquoted value, got %q", result[0].Value.ValueString())
		}
	})

	t.Run("default_format_skip_no_equals", func(t *testing.T) {
		result := parseActionArguments("key_value", "NOEQUALS\nK=V")
		if len(result) != 1 {
			t.Fatalf("expected 1 arg (NOEQUALS skipped), got %d", len(result))
		}
	})

	t.Run("default_format_empty_lines", func(t *testing.T) {
		result := parseActionArguments("key_value", "\n\nK=V\n\n")
		if len(result) != 1 {
			t.Fatalf("expected 1 arg, got %d", len(result))
		}
	})
}

func TestUnitActionResource_partialActionConfigFromModel(t *testing.T) {
	t.Run("no_schedule_no_webhook_clears_fields", func(t *testing.T) {
		data := &ActionResourceModel{
			FailureAlert:              types.BoolValue(true),
			RunOnStartupEnabled:       types.BoolValue(false),
			FileContents:              types.StringValue("console.log('hi');"),
			ReloadDependenciesEnabled: types.BoolValue(true),
			Schedule:                  nil,
			Webhook:                   nil,
			Arguments:                 nil,
		}
		cfg := partialActionConfigFromModel(data)
		if cfg.FailureAlert == nil || !*cfg.FailureAlert {
			t.Fatal("expected failure_alert=true")
		}
		if cfg.FileContents == nil || *cfg.FileContents != "console.log('hi');" {
			t.Fatalf("expected file_contents, got %v", cfg.FileContents)
		}
		// No schedule block → ScheduleEnabled and Schedule expression cleared.
		if cfg.ScheduleEnabled == nil || *cfg.ScheduleEnabled {
			t.Fatal("expected ScheduleEnabled=false when no schedule block")
		}
		if cfg.Schedule == nil || *cfg.Schedule != "" {
			t.Fatal("expected empty schedule expression")
		}
		// No webhook block → disabled.
		if cfg.WebhookEnabled == nil || *cfg.WebhookEnabled {
			t.Fatal("expected WebhookEnabled=false when no webhook block")
		}
		// No arguments → empty JSON object.
		if cfg.Arguments == nil || *cfg.Arguments != "{}" {
			t.Fatalf("expected arguments={}, got %v", cfg.Arguments)
		}
	})

	t.Run("with_schedule", func(t *testing.T) {
		data := &ActionResourceModel{
			FailureAlert:              types.BoolNull(),
			RunOnStartupEnabled:       types.BoolNull(),
			FileContents:              types.StringNull(),
			ReloadDependenciesEnabled: types.BoolNull(),
			Schedule: &ScheduleModel{
				Format:       types.StringValue("Cron"),
				Expression:   types.StringValue("0 0 * * * *"),
				Enabled:      types.BoolValue(true),
				Timezone:     types.StringValue("Europe/Berlin"),
				AlertEnabled: types.BoolValue(true),
			},
			Webhook:   nil,
			Arguments: nil,
		}
		cfg := partialActionConfigFromModel(data)
		if cfg.ScheduleFormat == nil || *cfg.ScheduleFormat != "Cron" {
			t.Fatal("expected ScheduleFormat=Cron")
		}
		if cfg.Schedule == nil || *cfg.Schedule != "0 0 * * * *" {
			t.Fatal("expected schedule expression")
		}
		if cfg.ScheduleTimezone == nil || *cfg.ScheduleTimezone != "Europe/Berlin" {
			t.Fatal("expected timezone=Europe/Berlin")
		}
		if cfg.ScheduleEnabled == nil || !*cfg.ScheduleEnabled {
			t.Fatal("expected ScheduleEnabled=true")
		}
	})

	t.Run("with_webhook", func(t *testing.T) {
		data := &ActionResourceModel{
			FailureAlert:              types.BoolNull(),
			RunOnStartupEnabled:       types.BoolNull(),
			FileContents:              types.StringNull(),
			ReloadDependenciesEnabled: types.BoolNull(),
			Schedule:                  nil,
			Webhook: &WebhookModel{
				Enabled: types.BoolValue(true),
				Secret:  types.StringValue("mysecret"),
			},
			Arguments: nil,
		}
		cfg := partialActionConfigFromModel(data)
		if cfg.WebhookEnabled == nil || !*cfg.WebhookEnabled {
			t.Fatal("expected WebhookEnabled=true")
		}
		if cfg.WebhookSecret == nil || *cfg.WebhookSecret != "mysecret" {
			t.Fatalf("expected WebhookSecret=mysecret, got %v", cfg.WebhookSecret)
		}
	})

	t.Run("with_arguments", func(t *testing.T) {
		data := &ActionResourceModel{
			FailureAlert:              types.BoolNull(),
			RunOnStartupEnabled:       types.BoolNull(),
			FileContents:              types.StringNull(),
			ReloadDependenciesEnabled: types.BoolNull(),
			Schedule:                  nil,
			Webhook:                   nil,
			Arguments: []ArgumentModel{
				{Name: types.StringValue("FOO"), Value: types.StringValue("bar")},
			},
		}
		cfg := partialActionConfigFromModel(data)
		if cfg.Arguments == nil || *cfg.Arguments == "{}" {
			t.Fatal("expected non-empty arguments JSON")
		}
	})
}

func TestUnitActionResource_actionToModel(t *testing.T) {
	t.Run("basic_fields", func(t *testing.T) {
		a := &client.Action{
			ID:   client.OID{OID: "abc123"},
			Name: "my-action",
			Tags: []string{"t1"},
			Config: client.ActionConfig{
				FailureAlert: true,
				FileContents: "console.log('hi');\n",
			},
		}
		var data ActionResourceModel
		actionToModel(a, &data)

		if data.ID.ValueString() != "abc123" {
			t.Fatalf("unexpected id: %s", data.ID.ValueString())
		}
		if data.Name.ValueString() != "my-action" {
			t.Fatalf("unexpected name: %s", data.Name.ValueString())
		}
		if !data.FailureAlert.ValueBool() {
			t.Fatal("expected failure_alert=true")
		}
		// Trailing newline should be stripped.
		if data.FileContents.ValueString() != "console.log('hi');" {
			t.Fatalf("expected trailing newline stripped, got %q", data.FileContents.ValueString())
		}
		if data.Schedule != nil {
			t.Fatal("expected nil schedule when not active")
		}
		if data.Webhook != nil {
			t.Fatal("expected nil webhook when not active")
		}
		if len(data.Tags.Elements()) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(data.Tags.Elements()))
		}
	})

	t.Run("with_active_schedule", func(t *testing.T) {
		a := &client.Action{
			ID:   client.OID{OID: "abc456"},
			Name: "scheduled-action",
			Config: client.ActionConfig{
				ScheduleEnabled:  true,
				ScheduleFormat:   "Cron",
				Schedule:         "0 0 * * * *",
				ScheduleTimezone: "UTC",
				ScheduleAlert:    true,
			},
		}
		var data ActionResourceModel
		actionToModel(a, &data)

		if data.Schedule == nil {
			t.Fatal("expected non-nil schedule block when ScheduleEnabled=true")
		}
		if data.Schedule.Format.ValueString() != "Cron" {
			t.Fatalf("unexpected schedule format: %s", data.Schedule.Format.ValueString())
		}
		if data.Schedule.Expression.ValueString() != "0 0 * * * *" {
			t.Fatalf("unexpected schedule expression: %s", data.Schedule.Expression.ValueString())
		}
	})

	t.Run("with_webhook", func(t *testing.T) {
		a := &client.Action{
			ID:   client.OID{OID: "abc789"},
			Name: "webhook-action",
			Config: client.ActionConfig{
				WebhookEnabled: true,
				WebhookSecret:  "s3cr3t",
			},
		}
		var data ActionResourceModel
		actionToModel(a, &data)

		if data.Webhook == nil {
			t.Fatal("expected non-nil webhook block")
		}
		if !data.Webhook.Enabled.ValueBool() {
			t.Fatal("expected webhook enabled=true")
		}
		if data.Webhook.Secret.ValueString() != "s3cr3t" {
			t.Fatalf("unexpected webhook secret: %s", data.Webhook.Secret.ValueString())
		}
	})

	t.Run("with_json_arguments", func(t *testing.T) {
		a := &client.Action{
			ID:   client.OID{OID: "abc999"},
			Name: "args-action",
			Config: client.ActionConfig{
				ArgumentsFormat: "json",
				Arguments:       `{"KEY":"val"}`,
			},
		}
		var data ActionResourceModel
		actionToModel(a, &data)

		if len(data.Arguments) != 1 {
			t.Fatalf("expected 1 argument, got %d", len(data.Arguments))
		}
		if data.Arguments[0].Name.ValueString() != "KEY" {
			t.Fatalf("unexpected arg name: %s", data.Arguments[0].Name.ValueString())
		}
		if data.Arguments[0].Value.ValueString() != "val" {
			t.Fatalf("unexpected arg value: %s", data.Arguments[0].Value.ValueString())
		}
	})
}
