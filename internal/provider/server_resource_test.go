// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestServerToResourceModel_PublicKeyUnknownResolvesToNull ensures that when the
// resource model carries an unknown public_key (as it does on a fresh Create where
// the plan holds "(known after apply)") the mapping function resolves it to null
// rather than leaving it unknown — which would cause Terraform to error with
// "provider still indicated an unknown value after apply".
func TestServerToResourceModel_PublicKeyUnknownResolvesToNull(t *testing.T) {
	ctx := context.Background()

	// Simulate the model as it arrives from the plan during a first Create:
	// public_key is unknown because no prior state exists for UseStateForUnknown()
	// to draw from and the user did not explicitly configure a value.
	data := &ServerResourceModel{
		PublicKey: types.StringUnknown(),
		// Initialise list/set fields that serverToResourceModel always writes so
		// that types.ListValueFrom calls on nil slices don't panic.
		Tags:              types.ListValueMust(types.StringType, nil),
		IgnoredDiskMounts: types.ListValueMust(types.StringType, nil),
		Links:             types.ListValueMust(types.StringType, nil),
	}

	server := &client.Server{
		ID:   client.OID{OID: "507f1f77bcf86cd799439011"},
		Name: "test-server",
		Tags: []string{},
		Config: client.ServerConfig{
			IgnoreMounts: []string{},
			Links:        []string{},
		},
	}

	diags := serverToResourceModel(ctx, server, data)
	if diags.HasError() {
		t.Fatalf("serverToResourceModel returned unexpected diagnostics: %v", diags)
	}

	if data.PublicKey.IsUnknown() {
		t.Fatal("public_key is still unknown after serverToResourceModel; " +
			"this would cause Terraform to error with 'provider still indicated an unknown value after apply'")
	}
	if !data.PublicKey.IsNull() {
		t.Fatalf("expected public_key to be null when not configured, got %q", data.PublicKey.ValueString())
	}
}

// TestServerToResourceModel_PublicKeyKnownValuePreserved ensures that a known
// public_key value written to state is not overwritten by serverToResourceModel
// (e.g. during Read or Update after the user explicitly set a key).
func TestServerToResourceModel_PublicKeyKnownValuePreserved(t *testing.T) {
	ctx := context.Background()

	const storedKey = "ssh-ed25519 AAAA..."

	data := &ServerResourceModel{
		PublicKey:         types.StringValue(storedKey),
		Tags:              types.ListValueMust(types.StringType, nil),
		IgnoredDiskMounts: types.ListValueMust(types.StringType, nil),
		Links:             types.ListValueMust(types.StringType, nil),
	}

	server := &client.Server{
		ID:   client.OID{OID: "507f1f77bcf86cd799439011"},
		Name: "test-server",
		Tags: []string{},
		Config: client.ServerConfig{
			IgnoreMounts: []string{},
			Links:        []string{},
		},
	}

	diags := serverToResourceModel(ctx, server, data)
	if diags.HasError() {
		t.Fatalf("serverToResourceModel returned unexpected diagnostics: %v", diags)
	}

	if data.PublicKey.IsUnknown() {
		t.Fatal("public_key must not become unknown")
	}
	if data.PublicKey.ValueString() != storedKey {
		t.Fatalf("expected public_key %q to be preserved, got %q", storedKey, data.PublicKey.ValueString())
	}
}

func TestAccServerResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfig("tf-acc-server-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-basic"),
					resource.TestCheckResourceAttrSet("komodo_server.test", "id"),
				),
			},
		},
	})
}

func TestAccServerResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfigWithRegion("tf-acc-server-update", "us-east"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-update"),
					resource.TestCheckResourceAttr("komodo_server.test", "region", "us-east"),
				),
			},
			{
				Config: testAccServerResourceConfigWithRegion("tf-acc-server-update", "us-west"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-update"),
					resource.TestCheckResourceAttr("komodo_server.test", "region", "us-west"),
				),
			},
		},
	})
}

func TestAccServerResource_importState(t *testing.T) {
	var serverID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfig("tf-acc-server-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_server.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_server.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						serverID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccServerResourceConfig("tf-acc-server-import"),
				ResourceName:      "komodo_server.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return serverID, nil },
			},
		},
	})
}

func TestAccServerResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfig("tf-acc-server-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_server.test", "id"),
					testAccServerDisappears("komodo_server.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServerResource_withAlertThresholds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfigWithThresholds("tf-acc-server-thresholds", 80.0, 95.0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-thresholds"),
					resource.TestCheckResourceAttr("komodo_server.test", "alerts.thresholds.cpu_warning", "80"),
					resource.TestCheckResourceAttr("komodo_server.test", "alerts.thresholds.cpu_critical", "95"),
				),
			},
		},
	})
}

func TestAccServerResource_withLinks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfigWithLinks("tf-acc-server-links", []string{"http://grafana.local", "http://prometheus.local"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-links"),
					resource.TestCheckResourceAttr("komodo_server.test", "links.#", "2"),
					resource.TestCheckResourceAttr("komodo_server.test", "links.0", "http://grafana.local"),
					resource.TestCheckResourceAttr("komodo_server.test", "links.1", "http://prometheus.local"),
				),
			},
		},
	})
}

func TestAccServerResource_rename(t *testing.T) {
	var savedID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfig("tf-acc-server-rename-orig"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-rename-orig"),
					resource.TestCheckResourceAttrSet("komodo_server.test", "id"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_server.test"]
						savedID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config: testAccServerResourceConfig("tf-acc-server-rename-new"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-rename-new"),
					func(s *terraform.State) error {
						rs := s.RootModule().Resources["komodo_server.test"]
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

func testAccServerDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteServer(context.Background(), rs.Primary.ID)
	}
}

func testAccServerResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name = %q
}
`, name)
}

func testAccServerResourceConfigWithRegion(name, region string) string {
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name   = %q
  region = %q
}
`, name, region)
}

func testAccServerResourceConfigWithThresholds(name string, cpuWarn, cpuCrit float64) string {
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name = %q
  alerts {
    thresholds {
      cpu_warning  = %g
      cpu_critical = %g
    }
  }
}
`, name, cpuWarn, cpuCrit)
}

func testAccServerResourceConfigWithLinks(name string, links []string) string {
	linksTF := "["
	for i, l := range links {
		if i > 0 {
			linksTF += ", "
		}
		linksTF += fmt.Sprintf("%q", l)
	}
	linksTF += "]"
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name  = %q
  links = %s
}
`, name, linksTF)
}

func TestAccServerResource_tags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerWithTagConfig("tf-acc-server-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "tags.#", "1"),
					resource.TestCheckResourceAttrPair("komodo_server.test", "tags.0", "komodo_tag.test", "id"),
				),
			},
			{
				Config: testAccServerClearTagsConfig("tf-acc-server-tags"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "tags.#", "0"),
				),
			},
		},
	})
}

func testAccServerWithTagConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_tag" "test" {
  name  = "tf-acc-tag-server"
  color = "Green"
}

resource "komodo_server" "test" {
  name = %q
  tags = [komodo_tag.test.id]
}
`, name)
}

func testAccServerClearTagsConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name = %q
  tags = []
}
`, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawServerPlan(t *testing.T, r *ServerResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawServerState(t *testing.T, r *ServerResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitServerResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &ServerResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitServerResource_createPlanGetError(t *testing.T) {
	r := &ServerResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawServerPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitServerResource_readStateGetError(t *testing.T) {
	r := &ServerResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawServerState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitServerResource_updatePlanGetError(t *testing.T) {
	r := &ServerResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawServerPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitServerResource_deleteStateGetError(t *testing.T) {
	r := &ServerResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawServerState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitServerResource_serverConfigFromModel(t *testing.T) {
	ctx := context.Background()

	t.Run("basic_fields", func(t *testing.T) {
		data := &ServerResourceModel{
			Tags:                              types.ListValueMust(types.StringType, nil),
			Address:                           types.StringValue("192.168.1.10:8120"),
			CertificateVerificationEnabled:    types.BoolValue(true),
			ExternalAddress:                   types.StringValue(""),
			Region:                            types.StringValue("eu-west"),
			PublicKey:                         types.StringNull(),
			Enabled:                           types.BoolValue(true),
			AutoRotateKeysEnabled:             types.BoolValue(false),
			AutoPruneImagesEnabled:            types.BoolValue(true),
			HistoricalSystemStatisticsEnabled: types.BoolValue(false),
			IgnoredDiskMounts:                 types.ListValueMust(types.StringType, nil),
			Links:                             types.ListValueMust(types.StringType, nil),
			Alerts:                            nil,
			Maintenance:                       nil,
		}
		cfg, diags := serverConfigFromModel(ctx, data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if cfg.Address == nil || *cfg.Address != "192.168.1.10:8120" {
			t.Fatalf("expected Address=192.168.1.10:8120, got %v", cfg.Address)
		}
		// InsecureTLS is the inverse of CertificateVerificationEnabled.
		if cfg.InsecureTLS == nil || *cfg.InsecureTLS {
			t.Fatal("expected InsecureTLS=false when certificate_verification_enabled=true")
		}
		if cfg.Region == nil || *cfg.Region != "eu-west" {
			t.Fatalf("expected Region=eu-west, got %v", cfg.Region)
		}
		if cfg.Enabled == nil || !*cfg.Enabled {
			t.Fatal("expected Enabled=true")
		}
		if cfg.AutoPrune == nil || !*cfg.AutoPrune {
			t.Fatal("expected AutoPrune=true")
		}
	})

	t.Run("alerts_enabled_with_cpu_and_disk", func(t *testing.T) {
		typesSet, _ := types.SetValueFrom(ctx, types.StringType, []string{"cpu", "disk"})
		data := &ServerResourceModel{
			Tags:                              types.ListValueMust(types.StringType, nil),
			Address:                           types.StringNull(),
			CertificateVerificationEnabled:    types.BoolNull(),
			ExternalAddress:                   types.StringNull(),
			Region:                            types.StringNull(),
			PublicKey:                         types.StringNull(),
			Enabled:                           types.BoolNull(),
			AutoRotateKeysEnabled:             types.BoolNull(),
			AutoPruneImagesEnabled:            types.BoolNull(),
			HistoricalSystemStatisticsEnabled: types.BoolNull(),
			IgnoredDiskMounts:                 types.ListValueMust(types.StringType, nil),
			Links:                             types.ListValueMust(types.StringType, nil),
			Alerts: &ServerAlertsModel{
				Enabled:    types.BoolValue(true),
				Types:      typesSet,
				Thresholds: nil,
			},
			Maintenance: nil,
		}
		cfg, diags := serverConfigFromModel(ctx, data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if cfg.SendCPUAlerts == nil || !*cfg.SendCPUAlerts {
			t.Fatal("expected SendCPUAlerts=true")
		}
		if cfg.SendDiskAlerts == nil || !*cfg.SendDiskAlerts {
			t.Fatal("expected SendDiskAlerts=true")
		}
		if cfg.SendMemAlerts == nil || *cfg.SendMemAlerts {
			t.Fatal("expected SendMemAlerts=false (not in types list)")
		}
		if cfg.SendUnreachableAlerts == nil || *cfg.SendUnreachableAlerts {
			t.Fatal("expected SendUnreachableAlerts=false")
		}
	})

	t.Run("alerts_disabled_clears_all_send_flags", func(t *testing.T) {
		emptySet, _ := types.SetValueFrom(ctx, types.StringType, []string{})
		data := &ServerResourceModel{
			Tags:                              types.ListValueMust(types.StringType, nil),
			Address:                           types.StringNull(),
			CertificateVerificationEnabled:    types.BoolNull(),
			ExternalAddress:                   types.StringNull(),
			Region:                            types.StringNull(),
			PublicKey:                         types.StringNull(),
			Enabled:                           types.BoolNull(),
			AutoRotateKeysEnabled:             types.BoolNull(),
			AutoPruneImagesEnabled:            types.BoolNull(),
			HistoricalSystemStatisticsEnabled: types.BoolNull(),
			IgnoredDiskMounts:                 types.ListValueMust(types.StringType, nil),
			Links:                             types.ListValueMust(types.StringType, nil),
			Alerts: &ServerAlertsModel{
				Enabled:    types.BoolValue(false),
				Types:      emptySet,
				Thresholds: nil,
			},
			Maintenance: nil,
		}
		cfg, diags := serverConfigFromModel(ctx, data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if cfg.SendCPUAlerts == nil || *cfg.SendCPUAlerts {
			t.Fatal("expected SendCPUAlerts=false when alerts.enabled=false")
		}
		if cfg.SendDiskAlerts == nil || *cfg.SendDiskAlerts {
			t.Fatal("expected SendDiskAlerts=false when alerts.enabled=false")
		}
		if cfg.SendMemAlerts == nil || *cfg.SendMemAlerts {
			t.Fatal("expected SendMemAlerts=false when alerts.enabled=false")
		}
	})

	t.Run("with_maintenance_window", func(t *testing.T) {
		data := &ServerResourceModel{
			Tags:                              types.ListValueMust(types.StringType, nil),
			Address:                           types.StringNull(),
			CertificateVerificationEnabled:    types.BoolNull(),
			ExternalAddress:                   types.StringNull(),
			Region:                            types.StringNull(),
			PublicKey:                         types.StringNull(),
			Enabled:                           types.BoolNull(),
			AutoRotateKeysEnabled:             types.BoolNull(),
			AutoPruneImagesEnabled:            types.BoolNull(),
			HistoricalSystemStatisticsEnabled: types.BoolNull(),
			IgnoredDiskMounts:                 types.ListValueMust(types.StringType, nil),
			Links:                             types.ListValueMust(types.StringType, nil),
			Alerts:                            nil,
			Maintenance: []MaintenanceWindowModel{
				{
					Name:            types.StringValue("weekly"),
					Description:     types.StringNull(),
					ScheduleType:    types.StringValue("Weekly"),
					DayOfWeek:       types.StringValue("Sunday"),
					Date:            types.StringNull(),
					Hour:            types.Int64Value(2),
					Minute:          types.Int64Value(30),
					DurationMinutes: types.Int64Value(60),
					Timezone:        types.StringValue("UTC"),
					Enabled:         types.BoolValue(true),
				},
			},
		}
		cfg, diags := serverConfigFromModel(ctx, data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if cfg.MaintenanceWindows == nil || len(*cfg.MaintenanceWindows) != 1 {
			t.Fatal("expected 1 maintenance window")
		}
		w := (*cfg.MaintenanceWindows)[0]
		if w.Name != "weekly" {
			t.Fatalf("expected window name=weekly, got %s", w.Name)
		}
		if w.ScheduleType != "Weekly" {
			t.Fatalf("expected schedule_type=Weekly, got %s", w.ScheduleType)
		}
		if w.Hour != 2 || w.Minute != 30 {
			t.Fatalf("unexpected hour/minute: %d:%d", w.Hour, w.Minute)
		}
		if !w.Enabled {
			t.Fatal("expected maintenance window enabled=true")
		}
	})
}

func TestUnitServerResource_serverToResourceModel(t *testing.T) {
	ctx := context.Background()

	t.Run("basic_fields_inverse_insecure_tls", func(t *testing.T) {
		s := &client.Server{
			ID:   client.OID{OID: "srv-abc"},
			Name: "my-server",
			Tags: []string{"prod"},
			Config: client.ServerConfig{
				Address:         "192.168.1.5:8120",
				InsecureTLS:     false,
				Region:          "eu-west",
				Enabled:         true,
				AutoRotateKeys:  false,
				AutoPrune:       true,
				StatsMonitoring: false,
			},
		}
		data := &ServerResourceModel{
			PublicKey: types.StringNull(),
		}
		diags := serverToResourceModel(ctx, s, data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if data.ID.ValueString() != "srv-abc" {
			t.Fatalf("unexpected ID: %s", data.ID.ValueString())
		}
		if data.Name.ValueString() != "my-server" {
			t.Fatalf("unexpected Name: %s", data.Name.ValueString())
		}
		if data.Address.ValueString() != "192.168.1.5:8120" {
			t.Fatalf("unexpected Address: %s", data.Address.ValueString())
		}
		// CertificateVerificationEnabled = !InsecureTLS = !false = true
		if !data.CertificateVerificationEnabled.ValueBool() {
			t.Fatal("expected CertificateVerificationEnabled=true when InsecureTLS=false")
		}
		if data.Region.ValueString() != "eu-west" {
			t.Fatalf("unexpected Region: %s", data.Region.ValueString())
		}
		if len(data.Tags.Elements()) != 1 {
			t.Fatalf("expected 1 tag, got %d", len(data.Tags.Elements()))
		}
	})

	t.Run("ignore_mounts_and_links_populated", func(t *testing.T) {
		s := &client.Server{
			ID:   client.OID{OID: "srv-mounts"},
			Name: "mount-server",
			Config: client.ServerConfig{
				IgnoreMounts: []string{"/dev", "/proc"},
				Links:        []string{"https://grafana.example.com"},
			},
		}
		data := &ServerResourceModel{PublicKey: types.StringNull()}
		diags := serverToResourceModel(ctx, s, data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if len(data.IgnoredDiskMounts.Elements()) != 2 {
			t.Fatalf("expected 2 ignored mounts, got %d", len(data.IgnoredDiskMounts.Elements()))
		}
		if len(data.Links.Elements()) != 1 {
			t.Fatalf("expected 1 link, got %d", len(data.Links.Elements()))
		}
	})

	t.Run("alerts_types_set_when_any_flag_enabled", func(t *testing.T) {
		s := &client.Server{
			ID:   client.OID{OID: "srv-alerts"},
			Name: "alert-server",
			Config: client.ServerConfig{
				SendCPUAlerts:  true,
				SendDiskAlerts: true,
				CPUCritical:    90.0,
				DiskWarning:    75.0,
			},
		}
		// Alerts block must be non-nil in prior model for the Alerts block to be populated.
		data := &ServerResourceModel{
			PublicKey: types.StringNull(),
			Alerts: &ServerAlertsModel{
				Enabled: types.BoolValue(true),
				Types:   types.SetValueMust(types.StringType, nil),
			},
		}
		diags := serverToResourceModel(ctx, s, data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if data.Alerts == nil {
			t.Fatal("expected non-nil alerts block")
		}
		var alertTypes []string
		_ = data.Alerts.Types.ElementsAs(ctx, &alertTypes, false)
		found := map[string]bool{}
		for _, a := range alertTypes {
			found[a] = true
		}
		if !found["cpu"] {
			t.Fatal("expected 'cpu' in alert types")
		}
		if !found["disk"] {
			t.Fatal("expected 'disk' in alert types")
		}
		if found["memory"] {
			t.Fatal("expected 'memory' not in alert types")
		}
		if data.Alerts.Thresholds == nil {
			t.Fatal("expected non-nil thresholds")
		}
		if data.Alerts.Thresholds.CPUCritical.ValueFloat64() != 90.0 {
			t.Fatalf("expected CPU critical=90.0, got %f", data.Alerts.Thresholds.CPUCritical.ValueFloat64())
		}
		if data.Alerts.Thresholds.DiskWarning.ValueFloat64() != 75.0 {
			t.Fatalf("expected disk warning=75.0, got %f", data.Alerts.Thresholds.DiskWarning.ValueFloat64())
		}
	})

	t.Run("maintenance_windows_populated", func(t *testing.T) {
		s := &client.Server{
			ID:   client.OID{OID: "srv-maint"},
			Name: "maint-server",
			Config: client.ServerConfig{
				MaintenanceWindows: []client.MaintenanceWindow{
					{
						Name:            "nightly",
						Description:     "",
						ScheduleType:    "Daily",
						Hour:            2,
						Minute:          0,
						DurationMinutes: 30,
						Timezone:        "",
						Enabled:         true,
					},
				},
			},
		}
		data := &ServerResourceModel{PublicKey: types.StringNull()}
		diags := serverToResourceModel(ctx, s, data)
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		if len(data.Maintenance) != 1 {
			t.Fatalf("expected 1 maintenance window, got %d", len(data.Maintenance))
		}
		w := data.Maintenance[0]
		if w.Name.ValueString() != "nightly" {
			t.Fatalf("expected window name=nightly, got %s", w.Name.ValueString())
		}
		if w.ScheduleType.ValueString() != "Daily" {
			t.Fatalf("expected schedule_type=Daily, got %s", w.ScheduleType.ValueString())
		}
		if w.Hour.ValueInt64() != 2 {
			t.Fatalf("expected hour=2, got %d", w.Hour.ValueInt64())
		}
		// Empty Description and Timezone should be mapped to null.
		if !w.Description.IsNull() {
			t.Fatal("expected null description for empty string")
		}
		if !w.Timezone.IsNull() {
			t.Fatal("expected null timezone for empty string")
		}
		if !w.Enabled.ValueBool() {
			t.Fatal("expected window enabled=true")
		}
	})
}

func TestUnitServerResource_validateConfig_alertsTypesEmptyWhenEnabled(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "komodo_server" "test" {
  name = "tf-test-server"
  alerts {
    enabled = true
    types   = []
  }
}`,
				ExpectError: regexp.MustCompile(`types must not be empty`),
			},
		},
	})
}

func TestUnitServerResource_validateConfig_alertsNilNoError(t *testing.T) {
	// Validator should early-return without error when alerts block is absent.
	r := &ServerResource{}
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	// Construct a minimal valid config (no alerts block)
	val := tftypes.NewValue(
		schemaResp.Schema.Type().TerraformType(ctx),
		nil, // null — triggers HasError in Get, but validator should not panic
	)
	req := fwresource.ValidateConfigRequest{Config: tfsdk.Config{Raw: val, Schema: schemaResp.Schema}}
	resp := &fwresource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)
	// null raw → Config.Get error, but validator should propagate that error gracefully
	// (it returns after HasError). We just verify no panic occurs.
	_ = resp.Diagnostics.HasError()
}

// ─── Mock-server unit tests ───────────────────────────────────────────────────

// mockValidServerJSON is a minimal but complete Server JSON that satisfies
// serverToResourceModel without diagnostics. ignore_mounts and links are non-null
// arrays so types.ListValueFrom never receives a nil slice.
const mockValidServerJSON = `{"_id":{"$oid":"507f1f77bcf86cd799439011"},"name":"tf-mock-server","tags":[],"config":{"address":"","insecure_tls":false,"external_address":"","region":"","enabled":false,"auto_rotate_keys":true,"passkey":"","auto_prune":false,"stats_monitoring":true,"ignore_mounts":[],"links":[]}}`

func TestUnitServerResource_createClientError(t *testing.T) {
	srv := newStatefulUserMockServer(t, func(typ string, _ int) (int, string) {
		if typ == "CreateServer" {
			return 500, `"create error"`
		}
		return 200, mockValidServerJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + `resource "komodo_server" "test" { name = "tf-mock-server" }`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestUnitServerResource_createMissingID(t *testing.T) {
	srv := newStatefulUserMockServer(t, func(typ string, _ int) (int, string) {
		if typ == "CreateServer" {
			return 200, `{"_id":{"$oid":""},"name":"tf-mock-server","tags":[],"config":{"ignore_mounts":[],"links":[]}}`
		}
		return 200, mockValidServerJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserProviderConfig(srv.URL) + `resource "komodo_server" "test" { name = "tf-mock-server" }`,
				ExpectError: regexp.MustCompile(`(?i)missing ID`),
			},
		},
	})
}

func TestUnitServerResource_deleteClientError(t *testing.T) {
	// Only the first DeleteServer call fails; subsequent cleanup calls succeed
	// so the framework's post-test destroy doesn't leave dangling resources.
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "DeleteServer" && n == 1 {
			return 500, `"delete error"`
		}
		return 200, mockValidServerJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + `resource "komodo_server" "test" { name = "tf-mock-server" }`,
			},
			{
				Config:      mockUserProviderConfig(srv.URL),
				Destroy:     true,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestUnitServerResource_updateRenameError(t *testing.T) {
	srv := newStatefulUserMockServer(t, func(typ string, _ int) (int, string) {
		if typ == "RenameServer" {
			return 500, `"rename error"`
		}
		return 200, mockValidServerJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + `resource "komodo_server" "test" { name = "tf-mock-server" }`,
			},
			{
				Config:      mockUserProviderConfig(srv.URL) + `resource "komodo_server" "test" { name = "tf-mock-server-v2" }`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestUnitServerResource_updateServerError(t *testing.T) {
	srv := newStatefulUserMockServer(t, func(typ string, _ int) (int, string) {
		if typ == "UpdateServer" {
			return 500, `"server update error"`
		}
		return 200, mockValidServerJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + `resource "komodo_server" "test" { name = "tf-mock-server" }`,
			},
			{
				// Same name (no rename), but explicit address triggers UpdateServer.
				Config: mockUserProviderConfig(srv.URL) + `
resource "komodo_server" "test" {
  name    = "tf-mock-server"
  address = "127.0.0.1:8120"
}`,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestUnitServerResource_updateGetAfterUpdateNil(t *testing.T) {
	// UpdateServer succeeds but the subsequent GetServer returns nil (not found).
	// n tracks per-type call count; the 3rd GetServer call is the post-update re-read.
	srv := newStatefulUserMockServer(t, func(typ string, n int) (int, string) {
		if typ == "GetServer" && n >= 3 {
			return 404, `"not found"`
		}
		if typ == "UpdateServer" {
			return 200, mockValidServerJSON
		}
		return 200, mockValidServerJSON
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserProviderConfig(srv.URL) + `resource "komodo_server" "test" { name = "tf-mock-server" }`,
			},
			{
				Config: mockUserProviderConfig(srv.URL) + `
resource "komodo_server" "test" {
  name    = "tf-mock-server"
  address = "127.0.0.1:8120"
}`,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}
