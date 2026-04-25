// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

func TestAccOnboardingKeyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-basic"),
					resource.TestCheckResourceAttrSet("komodo_onboarding_key.test", "public_key"),
					resource.TestCheckResourceAttrSet("komodo_onboarding_key.test", "private_key"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "enabled", "true"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "expires_at", ""),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "privileged", "false"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "create_builder", "false"),
				),
			},
		},
	})
}

func TestAccOnboardingKeyResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-update"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-update"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "enabled", "true"),
				),
			},
			// Update: disable and rename
			{
				Config: testAccOnboardingKeyResourceConfig_disabled("tf-onboarding-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-updated"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "enabled", "false"),
				),
			},
			// Re-enable
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-updated"),
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "enabled", "true"),
				),
			},
		},
	})
}

func TestAccOnboardingKeyResource_import(t *testing.T) {
	var publicKey string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and capture public_key
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-import"),
					resource.TestCheckResourceAttrSet("komodo_onboarding_key.test", "public_key"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_onboarding_key.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						publicKey = rs.Primary.Attributes["public_key"]
						return nil
					},
				),
			},
			// Import by public_key
			{
				Config:                               testAccOnboardingKeyResourceConfig_basic("tf-onboarding-import"),
				ResourceName:                         "komodo_onboarding_key.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "public_key",
				// private_key is only available on creation
				ImportStateVerifyIgnore: []string{"private_key"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return publicKey, nil
				},
			},
		},
	})
}

func TestAccOnboardingKeyResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOnboardingKeyResourceConfig_basic("tf-onboarding-disappear"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_onboarding_key.test", "name", "tf-onboarding-disappear"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccOnboardingKeyResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_onboarding_key" "test" {
  name = %[1]q
}
`, name)
}

func testAccOnboardingKeyResourceConfig_disabled(name string) string {
	return fmt.Sprintf(`
resource "komodo_onboarding_key" "test" {
  name    = %[1]q
  enabled = false
}
`, name)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawOnboardingKeyPlan(t *testing.T, r *OnboardingKeyResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawOnboardingKeyState(t *testing.T, r *OnboardingKeyResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitOnboardingKeyResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &OnboardingKeyResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitOnboardingKeyResource_createPlanGetError(t *testing.T) {
	r := &OnboardingKeyResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawOnboardingKeyPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitOnboardingKeyResource_readStateGetError(t *testing.T) {
	r := &OnboardingKeyResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawOnboardingKeyState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitOnboardingKeyResource_updatePlanGetError(t *testing.T) {
	r := &OnboardingKeyResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawOnboardingKeyPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitOnboardingKeyResource_deleteStateGetError(t *testing.T) {
	r := &OnboardingKeyResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawOnboardingKeyState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitOnboardingKey_versionAtLeast(t *testing.T) {
	cases := []struct {
		version string
		major   int
		minor   int
		patch   int
		want    bool
	}{
		// Exact match
		{"2.0.0", 2, 0, 0, true},
		{"1.5.3", 1, 5, 3, true},
		// Higher major
		{"3.0.0", 2, 0, 0, true},
		// Lower major
		{"1.9.9", 2, 0, 0, false},
		// Same major higher minor
		{"2.1.0", 2, 0, 0, true},
		// Same major lower minor
		{"2.0.0", 2, 1, 0, false},
		// Same major.minor higher patch
		{"2.0.5", 2, 0, 3, true},
		// Same major.minor lower patch
		{"2.0.2", 2, 0, 3, false},
		// Pre-release suffix is stripped — "2.0.0-beta" treated same as "2.0.0"
		{"2.0.0-beta", 2, 0, 0, true},
		// Build metadata suffix
		{"2.1.0+build42", 2, 1, 0, true},
		// Two-part version (missing patch defaults to 0)
		{"2.1", 2, 1, 0, true},
		// One-part version (missing minor/patch default to 0)
		{"2", 2, 0, 0, true},
		// Empty string
		{"", 1, 0, 0, false},
		// Non-numeric version
		{"invalid", 1, 0, 0, false},
	}
	for _, tc := range cases {
		name := fmt.Sprintf("%s>=%d.%d.%d", tc.version, tc.major, tc.minor, tc.patch)
		t.Run(name, func(t *testing.T) {
			got := versionAtLeast(tc.version, tc.major, tc.minor, tc.patch)
			if got != tc.want {
				t.Fatalf("versionAtLeast(%q, %d, %d, %d) = %v, want %v",
					tc.version, tc.major, tc.minor, tc.patch, got, tc.want)
			}
		})
	}
}
