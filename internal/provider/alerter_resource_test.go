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

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccAlerterResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-basic", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "name", "tf-acc-alerter-basic"),
					resource.TestCheckResourceAttr("komodo_alerter.test", "endpoint.type", "Custom"),
					resource.TestCheckResourceAttr("komodo_alerter.test", "endpoint.url", "http://localhost:7000"),
					resource.TestCheckResourceAttrSet("komodo_alerter.test", "id"),
				),
			},
		},
	})
}

func TestAccAlerterResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-update", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "endpoint.url", "http://localhost:7000"),
				),
			},
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-update", "http://localhost:8000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "endpoint.url", "http://localhost:8000"),
				),
			},
		},
	})
}

func TestAccAlerterResource_import(t *testing.T) {
	var alerterID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-import", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_alerter.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_alerter.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						alerterID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccAlerterResourceCustomConfig("tf-acc-alerter-import", "http://localhost:7000"),
				ResourceName:      "komodo_alerter.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return alerterID, nil },
			},
		},
	})
}

func TestAccAlerterResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-disappears", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_alerter.test", "id"),
					testAccAlerterDisappears("komodo_alerter.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAlerterDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteAlerter(context.Background(), rs.Primary.ID)
	}
}

func TestAccAlerterResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-rename-orig", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "name", "tf-acc-alerter-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_alerter.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_alerter.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccAlerterResourceCustomConfig("tf-acc-alerter-rename-new", "http://localhost:7000"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "name", "tf-acc-alerter-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_alerter.test"]
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

func testAccAlerterResourceCustomConfig(name, url string) string {
	return fmt.Sprintf(`
resource "komodo_alerter" "test" {
  name = %q
  endpoint {
    type = "Custom"
    url  = %q
  }
}
`, name, url)
}

func TestAccAlerterResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlerterWithTagConfig("tf-acc-alerter-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_alerter.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccAlerterClearTagsConfig("tf-acc-alerter-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_alerter.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccAlerterWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-alerter"
  color = "Green"
}

resource "komodo_alerter" "test" {
  name = %q
  endpoint {
    type = "Custom"
    url  = "http://localhost:7000"
  }
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccAlerterClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_alerter" "test" {
  name = %q
  endpoint {
    type = "Custom"
    url  = "http://localhost:7000"
  }
  tags = []
}
`, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawAlerterPlan(t *testing.T, r *AlerterResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawAlerterState(t *testing.T, r *AlerterResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitAlerterResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &AlerterResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitAlerterResource_createPlanGetError(t *testing.T) {
	r := &AlerterResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawAlerterPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitAlerterResource_readStateGetError(t *testing.T) {
	r := &AlerterResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawAlerterState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitAlerterResource_updatePlanGetError(t *testing.T) {
	r := &AlerterResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawAlerterPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitAlerterResource_deleteStateGetError(t *testing.T) {
	r := &AlerterResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawAlerterState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func makeAlerterWithEndpoint(epType string, params interface{}) *client.Alerter {
	b, _ := json.Marshal(params)
	return &client.Alerter{
		ID:   client.OID{OID: "test-id"},
		Name: "test",
		Tags: []string{},
		Config: client.AlerterConfig{
			Enabled:         true,
			AlertTypes:      []string{},
			Resources:       []client.ResourceTarget{},
			ExceptResources: []client.ResourceTarget{},
			Endpoint:        client.AlerterEndpoint{Type: epType, Params: json.RawMessage(b)},
		},
	}
}

func TestUnitAlerterResource_alerterToModel(t *testing.T) {
	ctx := context.Background()

	t.Run("slack", func(t *testing.T) {
		a := makeAlerterWithEndpoint("Slack", map[string]string{"url": "https://hooks.slack.com/test"})
		var data AlerterResourceModel
		diags := alerterToModel(ctx, a, &data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if data.Endpoint == nil || data.Endpoint.URL.ValueString() != "https://hooks.slack.com/test" {
			t.Fatal("expected Slack URL to be set")
		}
	})

	t.Run("discord", func(t *testing.T) {
		a := makeAlerterWithEndpoint("Discord", map[string]string{"url": "https://discord.com/api/webhooks/test"})
		var data AlerterResourceModel
		diags := alerterToModel(ctx, a, &data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if data.Endpoint.URL.ValueString() != "https://discord.com/api/webhooks/test" {
			t.Fatal("expected Discord URL")
		}
	})

	t.Run("ntfy_with_email", func(t *testing.T) {
		email := "alert@example.com"
		a := makeAlerterWithEndpoint("Ntfy", map[string]interface{}{"url": "https://ntfy.sh/test", "email": email})
		var data AlerterResourceModel
		diags := alerterToModel(ctx, a, &data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if data.Endpoint.Email.ValueString() != email {
			t.Fatalf("expected email %q, got %q", email, data.Endpoint.Email.ValueString())
		}
	})

	t.Run("ntfy_no_email", func(t *testing.T) {
		a := makeAlerterWithEndpoint("Ntfy", map[string]interface{}{"url": "https://ntfy.sh/test", "email": nil})
		var data AlerterResourceModel
		diags := alerterToModel(ctx, a, &data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if !data.Endpoint.Email.IsNull() {
			t.Fatal("expected null email for Ntfy without email")
		}
	})

	t.Run("pushover", func(t *testing.T) {
		a := makeAlerterWithEndpoint("Pushover", map[string]string{"url": "https://pushover.example/test"})
		var data AlerterResourceModel
		diags := alerterToModel(ctx, a, &data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if data.Endpoint.URL.ValueString() != "https://pushover.example/test" {
			t.Fatal("expected Pushover URL")
		}
	})

	t.Run("unknown_endpoint_type", func(t *testing.T) {
		a := makeAlerterWithEndpoint("Unknown", map[string]string{})
		var data AlerterResourceModel
		diags := alerterToModel(ctx, a, &data)
		if !diags.HasError() {
			t.Fatal("expected error for unknown endpoint type")
		}
	})

	t.Run("with_maintenance_windows", func(t *testing.T) {
		a := makeAlerterWithEndpoint("Slack", map[string]string{"url": "https://hooks.slack.com/test"})
		a.Config.MaintenanceWindows = []client.MaintenanceWindow{
			{
				Name:            "nightly",
				Description:     "Nightly window",
				ScheduleType:    "Daily",
				Hour:            2,
				Minute:          30,
				DurationMinutes: 60,
				Timezone:        "UTC",
				Enabled:         true,
			},
		}
		var data AlerterResourceModel
		diags := alerterToModel(ctx, a, &data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if len(data.Maintenance) != 1 || data.Maintenance[0].Name.ValueString() != "nightly" {
			t.Fatal("expected maintenance window to be set")
		}
	})

	t.Run("with_resources", func(t *testing.T) {
		a := makeAlerterWithEndpoint("Slack", map[string]string{"url": "https://hooks.slack.com/test"})
		a.Config.Resources = []client.ResourceTarget{{Type: "Stack", ID: "my-stack"}}
		a.Config.ExceptResources = []client.ResourceTarget{{Type: "Server", ID: "my-server"}}
		var data AlerterResourceModel
		diags := alerterToModel(ctx, a, &data)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if len(data.Resources) != 2 {
			t.Fatalf("expected 2 resource entries, got %d", len(data.Resources))
		}
	})
}

func TestUnitAlerterResource_alerterConfigInputFromModel(t *testing.T) {
	ctx := context.Background()

	makeModel := func(epType, url string) *AlerterResourceModel {
		return &AlerterResourceModel{
			Enabled:    types.BoolValue(true),
			AlertTypes: types.ListValueMust(types.StringType, nil),
			Endpoint:   &AlerterEndpointModel{Type: types.StringValue(epType), URL: types.StringValue(url), Email: types.StringNull()},
		}
	}

	t.Run("slack", func(t *testing.T) {
		m := makeModel("Slack", "https://hooks.slack.com/test")
		cfg, diags := alerterConfigInputFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if cfg.Endpoint == nil || cfg.Endpoint.Type != "Slack" {
			t.Fatal("expected Slack endpoint")
		}
	})

	t.Run("discord", func(t *testing.T) {
		m := makeModel("Discord", "https://discord.com/api/webhooks/test")
		cfg, diags := alerterConfigInputFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if cfg.Endpoint.Type != "Discord" {
			t.Fatal("expected Discord endpoint")
		}
	})

	t.Run("ntfy_with_email", func(t *testing.T) {
		m := makeModel("Ntfy", "https://ntfy.sh/test")
		m.Endpoint.Email = types.StringValue("alert@example.com")
		cfg, diags := alerterConfigInputFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if cfg.Endpoint.Type != "Ntfy" {
			t.Fatal("expected Ntfy endpoint")
		}
	})

	t.Run("ntfy_without_email", func(t *testing.T) {
		m := makeModel("Ntfy", "https://ntfy.sh/test")
		cfg, diags := alerterConfigInputFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if cfg.Endpoint.Type != "Ntfy" {
			t.Fatal("expected Ntfy endpoint")
		}
	})

	t.Run("pushover", func(t *testing.T) {
		m := makeModel("Pushover", "https://api.pushover.net/test")
		cfg, diags := alerterConfigInputFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if cfg.Endpoint.Type != "Pushover" {
			t.Fatal("expected Pushover endpoint")
		}
	})

	t.Run("unknown_type_error", func(t *testing.T) {
		m := makeModel("Unknown", "https://example.com")
		_, diags := alerterConfigInputFromModel(ctx, m)
		if !diags.HasError() {
			t.Fatal("expected error for unknown endpoint type")
		}
	})

	t.Run("nil_endpoint_error", func(t *testing.T) {
		m := &AlerterResourceModel{
			Enabled:    types.BoolValue(true),
			AlertTypes: types.ListValueMust(types.StringType, nil),
			Endpoint:   nil,
		}
		_, diags := alerterConfigInputFromModel(ctx, m)
		if !diags.HasError() {
			t.Fatal("expected error for nil endpoint")
		}
	})

	t.Run("with_maintenance_windows", func(t *testing.T) {
		m := makeModel("Slack", "https://hooks.slack.com/test")
		m.Maintenance = []MaintenanceWindowModel{
			{
				Name:            types.StringValue("nightly"),
				Description:     types.StringNull(),
				ScheduleType:    types.StringValue("Daily"),
				DayOfWeek:       types.StringNull(),
				Date:            types.StringNull(),
				Hour:            types.Int64Value(2),
				Minute:          types.Int64Value(0),
				DurationMinutes: types.Int64Value(60),
				Timezone:        types.StringNull(),
				Enabled:         types.BoolValue(true),
			},
		}
		cfg, diags := alerterConfigInputFromModel(ctx, m)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags)
		}
		if cfg.MaintenanceWindows == nil || len(*cfg.MaintenanceWindows) != 1 {
			t.Fatal("expected 1 maintenance window")
		}
	})
}

func wrongRawAlerterConfig(t *testing.T, r *AlerterResource) tfsdk.Config {
	t.Helper()
	schResp := &fwresource.SchemaResponse{}
	r.Schema(context.Background(), fwresource.SchemaRequest{}, schResp)
	return tfsdk.Config{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schResp.Schema,
	}
}

func TestUnitAlerterResource_validateConfig_configGetError(t *testing.T) {
	r := &AlerterResource{}
	req := fwresource.ValidateConfigRequest{Config: wrongRawAlerterConfig(t, r)}
	resp := &fwresource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed config")
	}
}
