// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	attr "github.com/hashicorp/terraform-plugin-framework/attr"
	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccProcedureResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "name", "tf-acc-procedure-basic"),
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
				),
			},
		},
	})
}

func TestAccProcedureResource_update(t *testing.T) {
	const name = "tf-acc-procedure-update"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "name", name),
				),
			},
			{
				Config: testAccProcedureResourceConfigWithSchedule(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.expression", "0 * * * *"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.enabled", "true"),
				),
			},
		},
	})
}

func TestAccProcedureResource_importState(t *testing.T) {
	var procedureID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_procedure.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						procedureID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccProcedureResourceConfig("tf-acc-procedure-import"),
				ResourceName:      "komodo_procedure.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return procedureID, nil
				},
				ImportStateVerifyIgnore: []string{"webhook"},
			},
		},
	})
}

func TestAccProcedureResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
					testAccProcedureDisappears("komodo_procedure.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccProcedureResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-rename-orig"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "name", "tf-acc-procedure-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_procedure.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-rename-new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "name", "tf-acc-procedure-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_procedure.test"]
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

func testAccProcedureDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteProcedure(context.Background(), rs.Primary.ID)
	}
}

func testAccProcedureResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
}
`, name)
}

func testAccProcedureResourceConfigWithSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
  schedule {
    format     = "Cron"
    expression = "0 * * * *"
    enabled    = true
  }
}
`, name)
}

func TestAccProcedureResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureWithTagConfig("tf-acc-procedure-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_procedure.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccProcedureClearTagsConfig("tf-acc-procedure-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccProcedureWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-procedure"
  color = "Green"
}

resource "komodo_procedure" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccProcedureClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
  tags = []
}
`, name)
}

func TestAccProcedureResource_scheduleAlertEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithFullSchedule("tf-acc-procedure-schedule-alert"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.expression", "0 * * * *"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.timezone", "Europe/Berlin"),
				),
			},
		},
	})
}

func TestAccProcedureResource_scheduleTimezone(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithFullSchedule("tf-acc-procedure-schedule-tz"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.timezone", "Europe/Berlin"),
				),
			},
		},
	})
}

func TestAccProcedureResource_scheduleDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithMinimalSchedule("tf-acc-procedure-schedule-defaults"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.expression", "0 * * * *"),
					// enabled and alert_enabled default to true; timezone defaults to ""
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.alert_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "schedule.timezone", ""),
				),
			},
		},
	})
}

func testAccProcedureResourceConfigWithFullSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
  schedule {
    format        = "Cron"
    expression    = "0 * * * *"
    enabled       = true
    alert_enabled = true
    timezone      = "Europe/Berlin"
  }
}
`, name)
}

func testAccProcedureResourceConfigWithMinimalSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name = %q
  schedule {
    format     = "Cron"
    expression = "0 * * * *"
  }
}
`, name)
}

// ---------------------------------------------------------------------------
// failure_alert_enabled
// ---------------------------------------------------------------------------

func TestAccProcedureResource_failureAlertDefault(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfig("tf-acc-procedure-failure-alert-default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// default is true
					resource.TestCheckResourceAttr("komodo_procedure.test", "failure_alert_enabled", "true"),
				),
			},
		},
	})
}

func TestAccProcedureResource_failureAlertDisabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithFailureAlert("tf-acc-procedure-failure-alert-off", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "failure_alert_enabled", "false"),
				),
			},
			{
				Config: testAccProcedureResourceConfigWithFailureAlert("tf-acc-procedure-failure-alert-off", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "failure_alert_enabled", "true"),
				),
			},
		},
	})
}

func testAccProcedureResourceConfigWithFailureAlert(name string, enabled bool) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "test" {
  name                  = %q
  failure_alert_enabled = %t
}
`, name, enabled)
}

// ---------------------------------------------------------------------------
// stage / execution blocks (native HCL, no jsonencode)
// ---------------------------------------------------------------------------

func TestAccProcedureResource_stages(t *testing.T) {
	const name = "tf-acc-procedure-stages"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with one stage containing one execution
			{
				Config: testAccProcedureResourceConfigWithStage(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.#", "1"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.name", "Deploy"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.#", "1"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.0.type", "RunProcedure"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.0.enabled", "true"),
				),
			},
			// Update to two stages
			{
				Config: testAccProcedureResourceConfigWithTwoStages(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.#", "2"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.name", "Stage1"),
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.1.name", "Stage2"),
				),
			},
			// Remove all stages
			{
				Config: testAccProcedureResourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.#", "0"),
				),
			},
		},
	})
}

func TestAccProcedureResource_executionEnabled(t *testing.T) {
	const name = "tf-acc-procedure-exec-enabled"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithDisabledExecution(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccProcedureResource_stagesImport(t *testing.T) {
	var procedureID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithStage("tf-acc-procedure-stages-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_procedure.test"]
						procedureID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccProcedureResourceConfigWithStage("tf-acc-procedure-stages-import"),
				ResourceName:      "komodo_procedure.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return procedureID, nil
				},
				ImportStateVerifyIgnore: []string{"webhook"},
			},
		},
	})
}

// ---------------------------------------------------------------------------
// execution parameters (native HCL map, no jsonencode)
// ---------------------------------------------------------------------------

func TestAccProcedureResource_executionParameters(t *testing.T) {
	const name = "tf-acc-procedure-exec-params"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureResourceConfigWithParametersV1(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_procedure.test", "stage.0.execution.0.parameters.%", "1"),
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "stage.0.execution.0.parameters.procedure"),
				),
			},
			// Update parameter value — should be detected as a change (drift detection)
			{
				Config: testAccProcedureResourceConfigWithParametersV2(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_procedure.test", "stage.0.execution.0.parameters.procedure"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Config helpers
// ---------------------------------------------------------------------------

func testAccProcedureResourceConfigWithStage(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child" {
  name = "%s-child"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Deploy"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }
}
`, name, name)
}

func testAccProcedureResourceConfigWithTwoStages(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child" {
  name = "%s-child"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Stage1"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }

  stage {
    name = "Stage2"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }
}
`, name, name)
}

func testAccProcedureResourceConfigWithDisabledExecution(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child" {
  name = "%s-child"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Stage1"

    execution {
      type    = "RunProcedure"
      enabled = false
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }
}
`, name, name)
}

func testAccProcedureResourceConfigWithParametersV1(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child1" {
  name = "%s-child1"
}

resource "komodo_procedure" "child2" {
  name = "%s-child2"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Run"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child1.id
      }
    }
  }
}
`, name, name, name)
}

func testAccProcedureResourceConfigWithParametersV2(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child1" {
  name = "%s-child1"
}

resource "komodo_procedure" "child2" {
  name = "%s-child2"
}

resource "komodo_procedure" "test" {
  name = %q

  stage {
    name = "Run"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child2.id
      }
    }
  }
}
`, name, name, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawProcedurePlan(t *testing.T, r *ProcedureResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawProcedureState(t *testing.T, r *ProcedureResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitProcedureResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &ProcedureResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitProcedureResource_createPlanGetError(t *testing.T) {
	r := &ProcedureResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawProcedurePlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitProcedureResource_readStateGetError(t *testing.T) {
	r := &ProcedureResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawProcedureState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitProcedureResource_updatePlanGetError(t *testing.T) {
	r := &ProcedureResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawProcedurePlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitProcedureResource_deleteStateGetError(t *testing.T) {
	r := &ProcedureResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawProcedureState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitProcedureResource_partialProcedureConfigFromModel(t *testing.T) {
	t.Run("empty_stages_no_schedule_no_webhook", func(t *testing.T) {
		data := &ProcedureResourceModel{
			FailureAlert: types.BoolValue(true),
			Stages:       nil,
			Schedule:     nil,
			Webhook:      nil,
		}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if len(cfg.Stages) != 0 {
			t.Fatalf("expected empty stages slice, got %d", len(cfg.Stages))
		}
		// No schedule block → ScheduleEnabled cleared.
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
		if cfg.FailureAlert == nil || !*cfg.FailureAlert {
			t.Fatal("expected failure_alert=true")
		}
	})

	t.Run("with_stages", func(t *testing.T) {
		data := &ProcedureResourceModel{
			FailureAlert: types.BoolNull(),
			Stages: []ProcedureStageModel{
				{
					Name: types.StringValue("stage-1"),
					Executions: []ProcedureExecutionModel{
						{
							Enabled:    types.BoolValue(true),
							Type:       types.StringValue("RunProcedure"),
							Parameters: types.MapNull(types.StringType),
						},
					},
				},
			},
			Schedule: nil,
			Webhook:  nil,
		}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if len(cfg.Stages) != 1 {
			t.Fatalf("expected 1 stage, got %d", len(cfg.Stages))
		}
		if cfg.Stages[0].Name != "stage-1" {
			t.Fatalf("unexpected stage name: %s", cfg.Stages[0].Name)
		}
		if len(cfg.Stages[0].Executions) != 1 {
			t.Fatal("expected 1 execution")
		}
		if !cfg.Stages[0].Executions[0].Enabled {
			t.Fatal("expected execution enabled=true")
		}
		if cfg.Stages[0].Executions[0].Execution.Type != "RunProcedure" {
			t.Fatalf("unexpected execution type: %s", cfg.Stages[0].Executions[0].Execution.Type)
		}
	})

	t.Run("with_schedule_and_webhook", func(t *testing.T) {
		data := &ProcedureResourceModel{
			FailureAlert: types.BoolValue(false),
			Stages:       nil,
			Schedule: &ScheduleModel{
				Format:       types.StringValue("Cron"),
				Expression:   types.StringValue("0 30 9 * * *"),
				Enabled:      types.BoolValue(true),
				Timezone:     types.StringValue("UTC"),
				AlertEnabled: types.BoolValue(false),
			},
			Webhook: &WebhookModel{
				Enabled: types.BoolValue(true),
				Secret:  types.StringNull(),
			},
		}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if cfg.ScheduleFormat == nil || *cfg.ScheduleFormat != "Cron" {
			t.Fatal("expected ScheduleFormat=Cron")
		}
		if cfg.Schedule == nil || *cfg.Schedule != "0 30 9 * * *" {
			t.Fatal("expected schedule expression")
		}
		if cfg.WebhookEnabled == nil || !*cfg.WebhookEnabled {
			t.Fatal("expected WebhookEnabled=true")
		}
	})
}

func TestUnitProcedureResource_procedureToModel(t *testing.T) {
	t.Run("empty_stages", func(t *testing.T) {
		proc := &client.Procedure{
			ID:   client.OID{OID: "proc123"},
			Name: "my-procedure",
			Tags: []string{},
			Config: client.ProcedureConfig{
				FailureAlert: true,
				Stages:       []client.ProcedureStage{},
			},
		}
		data := &ProcedureResourceModel{}
		procedureToModel(proc, data)

		if data.ID.ValueString() != "proc123" {
			t.Fatalf("unexpected id: %s", data.ID.ValueString())
		}
		if data.Name.ValueString() != "my-procedure" {
			t.Fatalf("unexpected name: %s", data.Name.ValueString())
		}
		if !data.FailureAlert.ValueBool() {
			t.Fatal("expected failure_alert=true")
		}
		if data.Stages != nil {
			t.Fatal("expected nil stages for empty API stages")
		}
		if data.Schedule != nil {
			t.Fatal("expected nil schedule when not active")
		}
	})

	t.Run("with_stage_no_prior_state", func(t *testing.T) {
		proc := &client.Procedure{
			ID:   client.OID{OID: "proc456"},
			Name: "staged-procedure",
			Config: client.ProcedureConfig{
				Stages: []client.ProcedureStage{
					{
						Name: "deploy",
						Executions: []client.ProcedureStageExecution{
							{
								Enabled: true,
								Execution: client.ProcedureExecution{
									Type:   "RunStack",
									Params: nil,
								},
							},
						},
					},
				},
			},
		}
		// No prior state (import path): data.Stages = nil.
		data := &ProcedureResourceModel{}
		procedureToModel(proc, data)

		if len(data.Stages) != 1 {
			t.Fatalf("expected 1 stage, got %d", len(data.Stages))
		}
		if data.Stages[0].Name.ValueString() != "deploy" {
			t.Fatalf("unexpected stage name: %s", data.Stages[0].Name.ValueString())
		}
		if len(data.Stages[0].Executions) != 1 {
			t.Fatal("expected 1 execution")
		}
		if data.Stages[0].Executions[0].Type.ValueString() != "RunStack" {
			t.Fatalf("unexpected execution type: %s", data.Stages[0].Executions[0].Type.ValueString())
		}
	})

	t.Run("with_active_schedule", func(t *testing.T) {
		proc := &client.Procedure{
			ID:   client.OID{OID: "proc789"},
			Name: "scheduled-procedure",
			Config: client.ProcedureConfig{
				ScheduleEnabled: true,
				ScheduleFormat:  "Cron",
				Schedule:        "0 0 * * * *",
			},
		}
		data := &ProcedureResourceModel{}
		procedureToModel(proc, data)

		if data.Schedule == nil {
			t.Fatal("expected non-nil schedule when ScheduleEnabled=true")
		}
		if data.Schedule.Format.ValueString() != "Cron" {
			t.Fatalf("unexpected schedule format: %s", data.Schedule.Format.ValueString())
		}
		if data.Schedule.Expression.ValueString() != "0 0 * * * *" {
			t.Fatalf("unexpected schedule expression: %s", data.Schedule.Expression.ValueString())
		}
	})

	t.Run("with_webhook_enabled", func(t *testing.T) {
		proc := &client.Procedure{
			ID:   client.OID{OID: "proc-wh"},
			Name: "webhook-procedure",
			Config: client.ProcedureConfig{
				WebhookEnabled: true,
				WebhookSecret:  "s3cr3t",
			},
		}
		data := &ProcedureResourceModel{}
		procedureToModel(proc, data)
		if data.Webhook == nil {
			t.Fatal("expected non-nil webhook block")
		}
		if !data.Webhook.Enabled.ValueBool() {
			t.Fatal("expected webhook enabled=true")
		}
		if data.Webhook.Secret.ValueString() != "s3cr3t" {
			t.Fatalf("expected webhook secret=s3cr3t, got %s", data.Webhook.Secret.ValueString())
		}
	})

	t.Run("with_stage_null_params_in_prior_state", func(t *testing.T) {
		// Prior state has a stage/execution with null parameters.
		// Expect: parameters stay null (existing.IsNull() branch).
		proc := &client.Procedure{
			ID:   client.OID{OID: "proc-null-params"},
			Name: "null-params-procedure",
			Config: client.ProcedureConfig{
				Stages: []client.ProcedureStage{
					{
						Name: "stage-1",
						Executions: []client.ProcedureStageExecution{
							{
								Enabled: true,
								Execution: client.ProcedureExecution{
									Type:   "RunAction",
									Params: nil,
								},
							},
						},
					},
				},
			},
		}
		// Prior state has a stage with null Parameters.
		data := &ProcedureResourceModel{
			Stages: []ProcedureStageModel{
				{
					Name: types.StringValue("stage-1"),
					Executions: []ProcedureExecutionModel{
						{
							Enabled:    types.BoolValue(true),
							Type:       types.StringValue("RunAction"),
							Parameters: types.MapNull(types.StringType),
						},
					},
				},
			},
		}
		procedureToModel(proc, data)
		if len(data.Stages) != 1 {
			t.Fatalf("expected 1 stage, got %d", len(data.Stages))
		}
		// Null parameters should stay null (not replaced with an empty map).
		if !data.Stages[0].Executions[0].Parameters.IsNull() {
			t.Fatal("expected parameters to remain null when prior was null")
		}
	})

	t.Run("with_stage_existing_params_merged_from_api", func(t *testing.T) {
		// Prior state has a stage with non-null parameters; API returns updated values.
		params := json.RawMessage(`{"procedure":"new-child","extra":"ignored"}`)
		proc := &client.Procedure{
			ID:   client.OID{OID: "proc-merge"},
			Name: "merge-params-procedure",
			Config: client.ProcedureConfig{
				Stages: []client.ProcedureStage{
					{
						Name: "run",
						Executions: []client.ProcedureStageExecution{
							{
								Enabled: true,
								Execution: client.ProcedureExecution{
									Type:   "RunProcedure",
									Params: params,
								},
							},
						},
					},
				},
			},
		}
		// Prior state has "procedure" key; "extra" is not in prior state so should be ignored.
		existingParams, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"procedure": types.StringValue("old-child"),
		})
		data := &ProcedureResourceModel{
			Stages: []ProcedureStageModel{
				{
					Name: types.StringValue("run"),
					Executions: []ProcedureExecutionModel{
						{
							Enabled:    types.BoolValue(true),
							Type:       types.StringValue("RunProcedure"),
							Parameters: existingParams,
						},
					},
				},
			},
		}
		procedureToModel(proc, data)
		if len(data.Stages) != 1 {
			t.Fatalf("expected 1 stage, got %d", len(data.Stages))
		}
		paramMap := data.Stages[0].Executions[0].Parameters
		if paramMap.IsNull() {
			t.Fatal("expected non-null parameters after merge")
		}
		// Only keys from prior state should be present; "extra" is excluded.
		if _, hasExtra := paramMap.Elements()["extra"]; hasExtra {
			t.Fatal("unexpected key 'extra' in merged parameters (not in prior state)")
		}
		val, ok := paramMap.Elements()["procedure"]
		if !ok {
			t.Fatal("expected key 'procedure' in merged parameters")
		}
		if val.(types.String).ValueString() != "new-child" {
			t.Fatalf("expected procedure=new-child from API, got %s", val.(types.String).ValueString())
		}
	})
}
func TestUnitProcedureResource_partialProcedureConfigFromModel_coercions(t *testing.T) {
	t.Run("empty_stages_sends_empty_slice", func(t *testing.T) {
		data := &ProcedureResourceModel{Stages: nil}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if cfg.Stages == nil {
			t.Fatal("expected non-nil Stages slice for empty stages")
		}
		if len(cfg.Stages) != 0 {
			t.Fatalf("expected 0 stages, got %d", len(cfg.Stages))
		}
	})

	t.Run("json_float64_coercion", func(t *testing.T) {
		// Numeric param "42" should be coerced to float64.
		params, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string{"stop_time": "42"})
		data := &ProcedureResourceModel{
			Stages: []ProcedureStageModel{
				{
					Name: types.StringValue("s1"),
					Executions: []ProcedureExecutionModel{
						{
							Enabled:    types.BoolValue(true),
							Type:       types.StringValue("RunBuild"),
							Parameters: params,
						},
					},
				},
			},
		}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if len(cfg.Stages) != 1 {
			t.Fatalf("expected 1 stage, got %d", len(cfg.Stages))
		}
	})

	t.Run("json_bool_coercion", func(t *testing.T) {
		params, _ := types.MapValueFrom(context.Background(), types.StringType, map[string]string{"enabled": "true"})
		data := &ProcedureResourceModel{
			Stages: []ProcedureStageModel{
				{
					Name: types.StringValue("s1"),
					Executions: []ProcedureExecutionModel{
						{
							Enabled:    types.BoolValue(true),
							Type:       types.StringValue("RunStack"),
							Parameters: params,
						},
					},
				},
			},
		}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		_ = cfg
	})

	t.Run("schedule_nil_clears_fields", func(t *testing.T) {
		data := &ProcedureResourceModel{Schedule: nil}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if cfg.ScheduleEnabled == nil || *cfg.ScheduleEnabled {
			t.Fatal("expected ScheduleEnabled=false when Schedule block is nil")
		}
		if cfg.Schedule == nil || *cfg.Schedule != "" {
			t.Fatal("expected empty Schedule when Schedule block is nil")
		}
	})

	t.Run("schedule_non_nil_sets_fields", func(t *testing.T) {
		data := &ProcedureResourceModel{
			Schedule: &ScheduleModel{
				Format:       types.StringValue("Cron"),
				Expression:   types.StringValue("0 * * * *"),
				Enabled:      types.BoolValue(true),
				Timezone:     types.StringValue("UTC"),
				AlertEnabled: types.BoolValue(false),
			},
		}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if cfg.ScheduleEnabled == nil || !*cfg.ScheduleEnabled {
			t.Fatal("expected ScheduleEnabled=true")
		}
		if cfg.Schedule == nil || *cfg.Schedule != "0 * * * *" {
			t.Fatalf("unexpected Schedule: %v", cfg.Schedule)
		}
	})

	t.Run("webhook_nil_clears_fields", func(t *testing.T) {
		data := &ProcedureResourceModel{Webhook: nil}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if cfg.WebhookEnabled == nil || *cfg.WebhookEnabled {
			t.Fatal("expected WebhookEnabled=false when Webhook block is nil")
		}
		if cfg.WebhookSecret == nil || *cfg.WebhookSecret != "" {
			t.Fatal("expected empty WebhookSecret when Webhook block is nil")
		}
	})

	t.Run("webhook_non_nil_sets_fields", func(t *testing.T) {
		data := &ProcedureResourceModel{
			Webhook: &WebhookModel{
				Enabled: types.BoolValue(true),
				Secret:  types.StringValue("mysecret"),
			},
		}
		cfg, diags := partialProcedureConfigFromModel(data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if cfg.WebhookEnabled == nil || !*cfg.WebhookEnabled {
			t.Fatal("expected WebhookEnabled=true")
		}
		if cfg.WebhookSecret == nil || *cfg.WebhookSecret != "mysecret" {
			t.Fatalf("unexpected WebhookSecret: %v", cfg.WebhookSecret)
		}
	})
}
