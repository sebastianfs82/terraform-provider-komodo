// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

// --- Unit tests ---

func TestSwarmToResourceModel_EmptyConfig(t *testing.T) {
	ctx := context.Background()

	data := &SwarmResourceModel{
		Tags:      types.ListValueMust(types.StringType, nil),
		ServerIDs: types.ListValueMust(types.StringType, nil),
		Links:     types.ListValueMust(types.StringType, nil),
	}

	swarm := &client.Swarm{
		ID:   client.OID{OID: "507f1f77bcf86cd799439011"},
		Name: "test-swarm",
		Tags: []string{},
		Config: client.SwarmConfig{
			ServerIDs:          []string{},
			Links:              []string{},
			AlertsEnabled:      true,
			MaintenanceWindows: nil,
		},
	}

	diags := swarmToResourceModel(ctx, swarm, data)
	if diags.HasError() {
		t.Fatalf("swarmToResourceModel returned unexpected diagnostics: %v", diags)
	}

	if data.ID.ValueString() != "507f1f77bcf86cd799439011" {
		t.Errorf("expected id %q, got %q", "507f1f77bcf86cd799439011", data.ID.ValueString())
	}
	if data.Name.ValueString() != "test-swarm" {
		t.Errorf("expected name %q, got %q", "test-swarm", data.Name.ValueString())
	}
	if !data.AlertsEnabled.ValueBool() {
		t.Error("expected alerts_enabled to be true")
	}
	if len(data.ServerIDs.Elements()) != 0 {
		t.Errorf("expected empty server_ids, got %d elements", len(data.ServerIDs.Elements()))
	}
	if data.Maintenance != nil {
		t.Errorf("expected nil maintenance, got %v", data.Maintenance)
	}
}

func TestSwarmToResourceModel_WithMaintenanceWindows(t *testing.T) {
	ctx := context.Background()

	data := &SwarmResourceModel{
		Tags:      types.ListValueMust(types.StringType, nil),
		ServerIDs: types.ListValueMust(types.StringType, nil),
		Links:     types.ListValueMust(types.StringType, nil),
	}

	swarm := &client.Swarm{
		ID:   client.OID{OID: "507f1f77bcf86cd799439012"},
		Name: "swarm-with-maintenance",
		Tags: []string{},
		Config: client.SwarmConfig{
			ServerIDs:     []string{},
			Links:         []string{},
			AlertsEnabled: false,
			MaintenanceWindows: []client.MaintenanceWindow{
				{
					Name:            "weekly-window",
					ScheduleType:    "Weekly",
					DayOfWeek:       "Sunday",
					Hour:            2,
					Minute:          0,
					DurationMinutes: 120,
					Enabled:         true,
				},
			},
		},
	}

	diags := swarmToResourceModel(ctx, swarm, data)
	if diags.HasError() {
		t.Fatalf("swarmToResourceModel returned unexpected diagnostics: %v", diags)
	}

	if len(data.Maintenance) != 1 {
		t.Fatalf("expected 1 maintenance window, got %d", len(data.Maintenance))
	}
	w := data.Maintenance[0]
	if w.Name.ValueString() != "weekly-window" {
		t.Errorf("expected maintenance name %q, got %q", "weekly-window", w.Name.ValueString())
	}
	if w.ScheduleType.ValueString() != "Weekly" {
		t.Errorf("expected schedule_type %q, got %q", "Weekly", w.ScheduleType.ValueString())
	}
	if w.DurationMinutes.ValueInt64() != 120 {
		t.Errorf("expected duration_minutes 120, got %d", w.DurationMinutes.ValueInt64())
	}
	if !w.Description.IsNull() {
		t.Errorf("expected description to be null, got %q", w.Description.ValueString())
	}
}

func TestSwarmConfigFromModel_AllFields(t *testing.T) {
	ctx := context.Background()

	serverIDs, _ := types.ListValueFrom(ctx, types.StringType, []string{"srv-1", "srv-2"})
	links, _ := types.ListValueFrom(ctx, types.StringType, []string{"http://monitor.local"})

	data := &SwarmResourceModel{
		ServerIDs:     serverIDs,
		Links:         links,
		AlertsEnabled: types.BoolValue(false),
		Maintenance:   nil,
	}

	cfg, diags := swarmConfigFromModel(ctx, data)
	if diags.HasError() {
		t.Fatalf("swarmConfigFromModel returned unexpected diagnostics: %v", diags)
	}

	if cfg.ServerIDs == nil || len(*cfg.ServerIDs) != 2 {
		t.Errorf("expected 2 server_ids, got %v", cfg.ServerIDs)
	}
	if cfg.Links == nil || len(*cfg.Links) != 1 || (*cfg.Links)[0] != "http://monitor.local" {
		t.Errorf("expected 1 link, got %v", cfg.Links)
	}
	if cfg.AlertsEnabled == nil || *cfg.AlertsEnabled != false {
		t.Errorf("expected alerts_enabled=false, got %v", cfg.AlertsEnabled)
	}
	if cfg.MaintenanceWindows != nil {
		t.Errorf("expected nil maintenance windows, got %v", cfg.MaintenanceWindows)
	}
}

// --- Acceptance tests ---

func TestAccSwarmResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmResourceConfig("tf-acc-swarm-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm.test", "name", "tf-acc-swarm-basic"),
					resource.TestCheckResourceAttrSet("komodo_swarm.test", "id"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "alerts_enabled", "true"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "server_ids.#", "0"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "links.#", "0"),
				),
			},
			// Verify plan is stable (no-diff on second apply).
			{
				Config:   testAccSwarmResourceConfig("tf-acc-swarm-basic"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccSwarmResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmResourceConfigWithAlerts("tf-acc-swarm-update", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm.test", "alerts_enabled", "true"),
				),
			},
			{
				Config: testAccSwarmResourceConfigWithAlerts("tf-acc-swarm-update", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm.test", "alerts_enabled", "false"),
				),
			},
		},
	})
}

func TestAccSwarmResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmResourceConfig("tf-acc-swarm-rename-orig"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm.test", "name", "tf-acc-swarm-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_swarm.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_swarm.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccSwarmResourceConfig("tf-acc-swarm-rename-new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm.test", "name", "tf-acc-swarm-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_swarm.test"]
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

func TestAccSwarmResource_withLinks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmResourceConfigWithLinks("tf-acc-swarm-links", []string{"http://grafana.local", "http://portainer.local"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm.test", "links.#", "2"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "links.0", "http://grafana.local"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "links.1", "http://portainer.local"),
				),
			},
		},
	})
}

func TestAccSwarmResource_withMaintenance(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmResourceConfigWithMaintenance("tf-acc-swarm-maint"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm.test", "name", "tf-acc-swarm-maint"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "maintenance.#", "1"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "maintenance.0.name", "weekly"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "maintenance.0.schedule_type", "Weekly"),
					resource.TestCheckResourceAttr("komodo_swarm.test", "maintenance.0.duration_minutes", "60"),
				),
			},
			{
				Config:   testAccSwarmResourceConfigWithMaintenance("tf-acc-swarm-maint"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccSwarmResource_importState(t *testing.T) {
	var swarmID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmResourceConfig("tf-acc-swarm-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_swarm.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_swarm.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						swarmID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccSwarmResourceConfig("tf-acc-swarm-import"),
				ResourceName:      "komodo_swarm.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return swarmID, nil },
			},
		},
	})
}

func TestAccSwarmResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmResourceConfig("tf-acc-swarm-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_swarm.test", "id"),
					testAccSwarmDisappears("komodo_swarm.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSwarmDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteSwarm(context.Background(), rs.Primary.ID)
	}
}

// --- Config helpers ---

func testAccSwarmResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_swarm" "test" {
  name = %q
}
`, name)
}

func testAccSwarmResourceConfigWithAlerts(name string, sendAlerts bool) string {
	return fmt.Sprintf(`
resource "komodo_swarm" "test" {
  name                  = %q
  alerts_enabled = %t
}
`, name, sendAlerts)
}

func testAccSwarmResourceConfigWithLinks(name string, links []string) string {
	linksHCL := "["
	for i, l := range links {
		if i > 0 {
			linksHCL += ", "
		}
		linksHCL += fmt.Sprintf("%q", l)
	}
	linksHCL += "]"
	return fmt.Sprintf(`
resource "komodo_swarm" "test" {
  name  = %q
  links = %s
}
`, name, linksHCL)
}

func testAccSwarmResourceConfigWithMaintenance(name string) string {
	return fmt.Sprintf(`
resource "komodo_swarm" "test" {
  name = %q

  maintenance {
    name             = "weekly"
    schedule_type    = "Weekly"
    day_of_week      = "Sunday"
    hour             = 2
    duration_minutes = 60
  }
}
`, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawSwarmPlan(t *testing.T, r *SwarmResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawSwarmState(t *testing.T, r *SwarmResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitSwarmResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &SwarmResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitSwarmResource_createPlanGetError(t *testing.T) {
	r := &SwarmResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawSwarmPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitSwarmResource_readStateGetError(t *testing.T) {
	r := &SwarmResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawSwarmState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitSwarmResource_updatePlanGetError(t *testing.T) {
	r := &SwarmResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawSwarmPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitSwarmResource_deleteStateGetError(t *testing.T) {
	r := &SwarmResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawSwarmState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestSwarmResource_swarmConfigFromModel_maintenanceWindows(t *testing.T) {
	ctx := context.Background()
	data := &SwarmResourceModel{
		ServerIDs:     types.ListValueMust(types.StringType, nil),
		Links:         types.ListValueMust(types.StringType, nil),
		AlertsEnabled: types.BoolNull(),
		Maintenance: []MaintenanceWindowModel{
			{
				Name:            types.StringValue("win1"),
				Description:     types.StringValue("desc"),
				ScheduleType:    types.StringValue("Daily"),
				DayOfWeek:       types.StringValue(""),
				Date:            types.StringValue(""),
				Hour:            types.Int64Value(2),
				Minute:          types.Int64Value(30),
				DurationMinutes: types.Int64Value(60),
				Timezone:        types.StringValue("UTC"),
				Enabled:         types.BoolValue(true),
			},
		},
	}
	cfg, diags := swarmConfigFromModel(ctx, data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.MaintenanceWindows == nil || len(*cfg.MaintenanceWindows) != 1 {
		t.Fatalf("expected 1 maintenance window, got %v", cfg.MaintenanceWindows)
	}
	w := (*cfg.MaintenanceWindows)[0]
	if w.Name != "win1" {
		t.Fatalf("unexpected Name: %s", w.Name)
	}
	if !w.Enabled {
		t.Fatal("expected window Enabled=true")
	}
}

func TestSwarmResource_swarmToResourceModel_maintenanceWindows(t *testing.T) {
	ctx := context.Background()

	t.Run("non_empty_windows_mapped", func(t *testing.T) {
		s := &client.Swarm{
			ID:   client.OID{OID: "swarm-maint"},
			Name: "test-swarm",
			Tags: []string{},
			Config: client.SwarmConfig{
				MaintenanceWindows: []client.MaintenanceWindow{
					{
						Name:            "win1",
						Description:     "",
						ScheduleType:    "Daily",
						DayOfWeek:       "",
						Date:            "",
						Hour:            2,
						Minute:          30,
						DurationMinutes: 60,
						Timezone:        "",
						Enabled:         true,
					},
				},
			},
		}
		data := &SwarmResourceModel{}
		diags := swarmToResourceModel(ctx, s, data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if len(data.Maintenance) != 1 {
			t.Fatalf("expected 1 maintenance window, got %d", len(data.Maintenance))
		}
		w := data.Maintenance[0]
		// Description empty → should be null
		if !w.Description.IsNull() {
			t.Fatalf("expected Description null, got %s", w.Description.ValueString())
		}
		// Date empty → should be null
		if !w.Date.IsNull() {
			t.Fatalf("expected Date null, got %s", w.Date.ValueString())
		}
		// Timezone empty → should be null
		if !w.Timezone.IsNull() {
			t.Fatalf("expected Timezone null, got %s", w.Timezone.ValueString())
		}
	})

	t.Run("nil_tags_handled", func(t *testing.T) {
		s := &client.Swarm{
			ID:     client.OID{OID: "swarm-notags"},
			Name:   "no-tags-swarm",
			Tags:   nil, // nil tags slice
			Config: client.SwarmConfig{},
		}
		data := &SwarmResourceModel{}
		diags := swarmToResourceModel(ctx, s, data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		if data.Tags.IsNull() || len(data.Tags.Elements()) != 0 {
			t.Fatalf("expected empty tags list, got %v", data.Tags)
		}
	})

	t.Run("empty_maintenance_with_prior_empty_keeps_empty_slice", func(t *testing.T) {
		s := &client.Swarm{
			ID:     client.OID{OID: "swarm-empty-maint"},
			Name:   "empty-maint",
			Tags:   []string{},
			Config: client.SwarmConfig{MaintenanceWindows: nil},
		}
		// Prior state has an empty (non-nil) slice.
		data := &SwarmResourceModel{
			Maintenance: []MaintenanceWindowModel{},
		}
		diags := swarmToResourceModel(ctx, s, data)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %v", diags)
		}
		// Should remain non-nil empty slice.
		if data.Maintenance == nil {
			t.Fatal("expected non-nil empty maintenance slice when prior was empty slice")
		}
		if len(data.Maintenance) != 0 {
			t.Fatalf("expected 0 maintenance windows, got %d", len(data.Maintenance))
		}
	})
}
