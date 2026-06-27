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
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"
)

func TestAccSwarmSecretResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmSecretResourceConfig("tf-acc-swarm-secret-basic", "tf-acc-secret-basic", "initial-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm_secret.test", "name", "tf-acc-secret-basic"),
					resource.TestCheckResourceAttr("komodo_swarm_secret.test", "swarm", "tf-acc-swarm-secret-basic"),
					resource.TestCheckResourceAttrSet("komodo_swarm_secret.test", "id"),
				),
			},
			{
				Config:   testAccSwarmSecretResourceConfig("tf-acc-swarm-secret-basic", "tf-acc-secret-basic", "initial-value"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccSwarmSecretResource_rotate(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmSecretResourceConfig("tf-acc-swarm-secret-rotate", "tf-acc-secret-rotate", "value-one"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm_secret.test", "name", "tf-acc-secret-rotate"),
				),
			},
			{
				Config: testAccSwarmSecretResourceConfig("tf-acc-swarm-secret-rotate", "tf-acc-secret-rotate", "value-two"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_swarm_secret.test", "name", "tf-acc-secret-rotate"),
				),
			},
		},
	})
}

func TestAccSwarmSecretResource_importState(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmSecretResourceConfig("tf-acc-swarm-secret-import", "tf-acc-secret-import", "import-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_swarm_secret.test", "id"),
				),
			},
			{
				Config:                  testAccSwarmSecretResourceConfig("tf-acc-swarm-secret-import", "tf-acc-secret-import", "import-value"),
				ResourceName:            "komodo_swarm_secret.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"data"},
			},
		},
	})
}

func TestAccSwarmSecretResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSwarmSecretResourceConfig("tf-acc-swarm-secret-disappears", "tf-acc-secret-disappears", "disappear-value"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_swarm_secret.test", "id"),
					testAccSwarmSecretDisappears("komodo_swarm_secret.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSwarmSecretDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found in state: %s", resourceName)
		}
		swarm := rs.Primary.Attributes["swarm"]
		name := rs.Primary.Attributes["name"]

		c := client.NewClient(
			os.Getenv("KOMODO_ENDPOINT"),
			os.Getenv("KOMODO_USERNAME"),
			os.Getenv("KOMODO_PASSWORD"),
		)
		return c.RemoveSwarmSecrets(context.Background(), client.RemoveSwarmSecretsRequest{
			Swarm:   swarm,
			Secrets: []string{name},
		})
	}
}

func testAccSwarmSecretResourceConfig(swarmName, secretName, data string) string {
	return fmt.Sprintf(`
resource "komodo_swarm" "test" {
  name = %q
}

resource "komodo_swarm_secret" "test" {
  swarm = komodo_swarm.test.name
  name  = %q
  data  = %q
}
`, swarmName, secretName, data)
}

func wrongRawSwarmSecretPlan(t *testing.T, r *SwarmSecretResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawSwarmSecretState(t *testing.T, r *SwarmSecretResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitSwarmSecretResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &SwarmSecretResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitSwarmSecretResource_createPlanGetError(t *testing.T) {
	r := &SwarmSecretResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawSwarmSecretPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitSwarmSecretResource_readStateGetError(t *testing.T) {
	r := &SwarmSecretResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawSwarmSecretState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitSwarmSecretResource_updatePlanGetError(t *testing.T) {
	r := &SwarmSecretResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawSwarmSecretPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitSwarmSecretResource_deleteStateGetError(t *testing.T) {
	r := &SwarmSecretResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawSwarmSecretState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitSwarmSecretResource_importState_invalidFormat(t *testing.T) {
	r := &SwarmSecretResource{}
	req := fwresource.ImportStateRequest{ID: "no-colon-separator"}
	resp := &fwresource.ImportStateResponse{}
	r.ImportState(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for invalid import ID format")
	}
}
