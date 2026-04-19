// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServiceUserDataSource_byUsername(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserDataSourceConfig_byUsername("tf-svc-ds-name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "username", "tf-svc-ds-name"),
					resource.TestCheckResourceAttrSet("data.komodo_service_user.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "admin_enabled", "false"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "create_server_enabled", "false"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "create_build_enabled", "false"),
				),
			},
		},
	})
}

func TestAccServiceUserDataSource_byID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserDataSourceConfig_byID("tf-svc-ds-id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "username", "tf-svc-ds-id"),
					resource.TestCheckResourceAttrPair(
						"data.komodo_service_user.test", "id",
						"komodo_service_user.test", "id",
					),
				),
			},
		},
	})
}

func TestAccServiceUserDataSource_withPermissions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUserDataSourceConfig_withPermissions("tf-svc-ds-perms", true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "username", "tf-svc-ds-perms"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "create_server_enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_service_user.test", "create_build_enabled", "true"),
				),
			},
		},
	})
}

// Config helpers

func testAccServiceUserDataSourceConfig_byUsername(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
}

data "komodo_service_user" "test" {
  username = komodo_service_user.test.username
}
`, username)
}

func testAccServiceUserDataSourceConfig_byID(username string) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username = %[1]q
}

data "komodo_service_user" "test" {
  id = komodo_service_user.test.id
}
`, username)
}

func testAccServiceUserDataSourceConfig_withPermissions(username string, createServers, createBuilds bool) string {
	return fmt.Sprintf(`
resource "komodo_service_user" "test" {
  username               = %[1]q
  create_server_enabled  = %[2]t
  create_build_enabled   = %[3]t
}

data "komodo_service_user" "test" {
  username = komodo_service_user.test.username
}
`, username, createServers, createBuilds)
}

func TestUnitServiceUserDataSource_configure(t *testing.T) {
	d := &ServiceUserDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}

func TestAccServiceUserDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceUserDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccServiceUserDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceUserDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccServiceUserDataSourceConfig_bothSet() string {
	return `
data "komodo_service_user" "test" {
  id       = "some-id"
  username = "some-username"
}
`
}

func testAccServiceUserDataSourceConfig_neitherSet() string {
	return `
data "komodo_service_user" "test" {}
`
}

// ─── Unit tests for ValidateConfig / Read HasError ────────────────────────────

func TestUnitServiceUserDataSource_validateConfigGetError(t *testing.T) {
	d := &ServiceUserDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	req := datasource.ValidateConfigRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(tftypes.String, "invalid"),
			Schema: schemaResp.Schema,
		},
	}
	resp := &datasource.ValidateConfigResponse{}
	d.ValidateConfig(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error from malformed config")
	}
}

func TestUnitServiceUserDataSource_validateConfigUnknown(t *testing.T) {
	// username is Unknown → ValidateConfig should return early without error.
	d := &ServiceUserDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	objType := schemaResp.Schema.Type().TerraformType(ctx)
	configVal := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                    tftypes.NewValue(tftypes.String, nil),
		"username":              tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"enabled":               tftypes.NewValue(tftypes.Bool, nil),
		"admin_enabled":         tftypes.NewValue(tftypes.Bool, nil),
		"create_server_enabled": tftypes.NewValue(tftypes.Bool, nil),
		"create_build_enabled":  tftypes.NewValue(tftypes.Bool, nil),
	})
	req := datasource.ValidateConfigRequest{
		Config: tfsdk.Config{Raw: configVal, Schema: schemaResp.Schema},
	}
	resp := &datasource.ValidateConfigResponse{}
	d.ValidateConfig(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected no error for unknown username: %v", resp.Diagnostics)
	}
}

func TestUnitServiceUserDataSource_readConfigGetError(t *testing.T) {
	d := &ServiceUserDataSource{}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	req := datasource.ReadRequest{
		Config: tfsdk.Config{
			Raw:    tftypes.NewValue(tftypes.String, "invalid"),
			Schema: schemaResp.Schema,
		},
	}
	resp := &datasource.ReadResponse{}
	d.Read(ctx, req, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error from malformed config")
	}
}

// ─── Mock-server tests ────────────────────────────────────────────────────────

const mockSvcUserDataSourceConfig = `
data "komodo_service_user" "test" {
  username = "tf-mock-svc-user"
}
`

func TestAccServiceUserDataSource_findUserError(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusInternalServerError, `"find error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserDataSourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccServiceUserDataSource_findUserNil(t *testing.T) {
	srv := newSvcUserMockServer(t, map[string]mockUserRoute{
		"FindUser": {http.StatusNotFound, `"not found"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockSvcUserProviderConfig(srv.URL) + mockSvcUserDataSourceConfig,
				ExpectError: regexp.MustCompile(`(?i)not found`),
			},
		},
	})
}

// ─── Successful data source read test ────────────────────────────────────────

func TestAccServiceUserDataSource_readSuccess(t *testing.T) {
	srv := newSvcUserMockServer(t, nil)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockSvcUserProviderConfig(srv.URL) + mockSvcUserDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.komodo_service_user.test", "username", "tf-mock-svc-user"),
			},
		},
	})
}
