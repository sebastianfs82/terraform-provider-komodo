// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccRegistryAccountResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig("registry.example.com", "testuser", "mytoken123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_registry_account.test", "domain", "registry.example.com"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "username", "testuser"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "token", "mytoken123"),
					resource.TestCheckResourceAttrSet("komodo_registry_account.test", "id"),
				),
			},
		},
	})
}

func TestAccRegistryAccountResource_updateToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig("ghcr.io", "updateuser", "original-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_registry_account.test", "domain", "ghcr.io"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "username", "updateuser"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "token", "original-token"),
				),
			},
			{
				Config: testAccRegistryAccountResourceConfig("ghcr.io", "updateuser", "updated-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_registry_account.test", "domain", "ghcr.io"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "username", "updateuser"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "token", "updated-token"),
				),
			},
		},
	})
}

func TestAccRegistryAccountResource_import(t *testing.T) {
	var accountID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig("registry.example.com", "importuser", "import-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_registry_account.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_registry_account.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						accountID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:                  testAccRegistryAccountResourceConfig("registry.example.com", "importuser", "import-token"),
				ResourceName:            "komodo_registry_account.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       func(_ *terraform.State) (string, error) { return accountID, nil },
				ImportStateVerifyIgnore: []string{"token"},
			},
		},
	})
}

func TestAccRegistryAccountResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig("registry.example.com", "disappearuser", "disappear-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_registry_account.test", "id"),
					testAccRegistryAccountDisappears("komodo_registry_account.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRegistryAccountDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteDockerRegistryAccount(context.Background(), rs.Primary.ID)
	}
}

func TestAccRegistryAccountResource_noToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig_noToken("registry.example.com", "notokenuser"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_registry_account.test", "domain", "registry.example.com"),
					resource.TestCheckResourceAttr("komodo_registry_account.test", "username", "notokenuser"),
					resource.TestCheckNoResourceAttr("komodo_registry_account.test", "token"),
					resource.TestCheckResourceAttrSet("komodo_registry_account.test", "id"),
				),
			},
		},
	})
}

func TestAccRegistryAccountResource_addToken(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRegistryAccountResourceConfig_noToken("registry.example.com", "addtokenuser"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_registry_account.test", "token"),
				),
			},
			{
				Config: testAccRegistryAccountResourceConfig("registry.example.com", "addtokenuser", "newly-added-token"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_registry_account.test", "token", "newly-added-token"),
				),
			},
		},
	})
}

func testAccRegistryAccountResourceConfig_noToken(domain, username string) string {
	return fmt.Sprintf(`
resource "komodo_registry_account" "test" {
  domain   = %q
  username = %q
}
`, domain, username)
}

func testAccRegistryAccountResourceConfig(domain, username, token string) string {
	return fmt.Sprintf(`
resource "komodo_registry_account" "test" {
  domain   = %q
  username = %q
  token    = %q
}
`, domain, username, token)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawRegistryAccountPlan(t *testing.T, r *RegistryAccountResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawRegistryAccountState(t *testing.T, r *RegistryAccountResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitRegistryAccountResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &RegistryAccountResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitRegistryAccountResource_createPlanGetError(t *testing.T) {
	r := &RegistryAccountResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawRegistryAccountPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitRegistryAccountResource_readStateGetError(t *testing.T) {
	r := &RegistryAccountResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawRegistryAccountState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitRegistryAccountResource_updatePlanGetError(t *testing.T) {
	r := &RegistryAccountResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawRegistryAccountPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitRegistryAccountResource_deleteStateGetError(t *testing.T) {
	r := &RegistryAccountResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawRegistryAccountState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}
