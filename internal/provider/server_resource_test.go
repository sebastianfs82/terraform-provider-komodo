// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/types"
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
  alerts = {
    thresholds = {
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
