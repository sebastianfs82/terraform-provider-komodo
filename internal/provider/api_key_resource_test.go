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

func TestAccApiKeyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfig("tf-acc-api-key-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_api_key.test", "name", "tf-acc-api-key-basic"),
					resource.TestCheckResourceAttr("komodo_api_key.test", "expires_at", ""),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "secret"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "user_id"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "created_at"),
				),
			},
		},
	})
}

func TestAccApiKeyResource_withExpiration(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfigWithExpiration("tf-acc-api-key-expiring", "2030-01-01T00:00:00Z"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_api_key.test", "name", "tf-acc-api-key-expiring"),
					resource.TestCheckResourceAttr("komodo_api_key.test", "expires_at", "2030-01-01T00:00:00Z"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "secret"),
				),
			},
		},
	})
}

func TestAccApiKeyResource_serviceUser(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfigServiceUser("tf-acc-svc-apikey", "tf-acc-svc-apikey-key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_api_key.test", "name", "tf-acc-svc-apikey-key"),
					resource.TestCheckResourceAttr("komodo_api_key.test", "expires_at", ""),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "secret"),
					resource.TestCheckResourceAttrPair(
						"komodo_api_key.test", "service_user_id",
						"komodo_service_user.svc", "id",
					),
				),
			},
		},
	})
}

func TestAccApiKeyResource_importState(t *testing.T) {
	var keyValue string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfig("tf-acc-api-key-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_api_key.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						keyValue = rs.Primary.Attributes["key"]
						return nil
					},
				),
			},
			{
				Config:            testAccApiKeyResourceConfig("tf-acc-api-key-import"),
				ResourceName:      "komodo_api_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Secret is only available on creation and cannot be retrieved via import.
				ImportStateVerifyIgnore: []string{"secret"},
				ImportStateIdFunc: func(_ *terraform.State) (string, error) {
					return keyValue, nil
				},
			},
		},
	})
}

func TestAccApiKeyResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeyResourceConfig("tf-acc-api-key-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_api_key.test", "key"),
					testAccApiKeyDisappears("komodo_api_key.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccApiKeyDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteApiKey(context.Background(), client.DeleteApiKeyRequest{Key: rs.Primary.Attributes["key"]})
	}
}

func testAccApiKeyResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_api_key" "test" {
  name = %q
}
`, name)
}

func testAccApiKeyResourceConfigWithExpiration(name string, expires string) string {
	return fmt.Sprintf(`
resource "komodo_api_key" "test" {
  name    = %q
  expires_at = %q
}
`, name, expires)
}

func testAccApiKeyResourceConfigServiceUser(username, keyName string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "svc" {
  username = %q
}

resource "komodo_api_key" "test" {
  name            = %q
  service_user_id = komodo_service_user.svc.id
}
`, username, keyName)
}

// ─── Unit tests ──────────────────────────────────────────────────────────────

func wrongRawApiKeyPlan(t *testing.T, r *ApiKeyResource) tfsdk.Plan {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.Plan{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func wrongRawApiKeyState(t *testing.T, r *ApiKeyResource) tfsdk.State {
	t.Helper()
	ctx := context.Background()
	schemaResp := &fwresource.SchemaResponse{}
	r.Schema(ctx, fwresource.SchemaRequest{}, schemaResp)
	return tfsdk.State{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schemaResp.Schema,
	}
}

func TestUnitApiKeyResource_configure(t *testing.T) {
	t.Run("wrong_type", func(t *testing.T) {
		r := &ApiKeyResource{}
		req := fwresource.ConfigureRequest{ProviderData: "not-a-client"}
		resp := &fwresource.ConfigureResponse{}
		r.Configure(context.Background(), req, resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected diagnostic error for wrong ProviderData type")
		}
	})
}

func TestUnitApiKeyResource_createPlanGetError(t *testing.T) {
	r := &ApiKeyResource{client: &client.Client{}}
	req := fwresource.CreateRequest{Plan: wrongRawApiKeyPlan(t, r)}
	resp := &fwresource.CreateResponse{}
	r.Create(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func TestUnitApiKeyResource_readStateGetError(t *testing.T) {
	r := &ApiKeyResource{client: &client.Client{}}
	req := fwresource.ReadRequest{State: wrongRawApiKeyState(t, r)}
	resp := &fwresource.ReadResponse{}
	r.Read(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitApiKeyResource_deleteStateGetError(t *testing.T) {
	r := &ApiKeyResource{client: &client.Client{}}
	req := fwresource.DeleteRequest{State: wrongRawApiKeyState(t, r)}
	resp := &fwresource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed state")
	}
}

func TestUnitApiKeyResource_updatePlanGetError(t *testing.T) {
	r := &ApiKeyResource{client: &client.Client{}}
	req := fwresource.UpdateRequest{Plan: wrongRawApiKeyPlan(t, r)}
	resp := &fwresource.UpdateResponse{}
	r.Update(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed plan")
	}
}

func wrongRawApiKeyConfig(t *testing.T, r *ApiKeyResource) tfsdk.Config {
	t.Helper()
	schResp := &fwresource.SchemaResponse{}
	r.Schema(context.Background(), fwresource.SchemaRequest{}, schResp)
	return tfsdk.Config{
		Raw:    tftypes.NewValue(tftypes.String, "invalid"),
		Schema: schResp.Schema,
	}
}

func TestUnitApiKeyResource_validateConfig_configGetError(t *testing.T) {
	r := &ApiKeyResource{}
	req := fwresource.ValidateConfigRequest{Config: wrongRawApiKeyConfig(t, r)}
	resp := &fwresource.ValidateConfigResponse{}
	r.ValidateConfig(context.Background(), req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for malformed config")
	}
}
