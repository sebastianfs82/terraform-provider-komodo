// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserGroupsDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupsDataSourceConfig_basic("tf-test-groups-ds"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// At least one group must exist (the one we created)
					resource.TestCheckResourceAttrSet("data.komodo_user_groups.test", "groups.#"),
				),
			},
		},
	})
}

func TestAccUserGroupsDataSource_containsCreatedGroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserGroupsDataSourceConfig_basic("tf-test-groups-ds-find"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.komodo_user_groups.test",
						"groups.*",
						map[string]string{
							"name": "tf-test-groups-ds-find",
						},
					),
				),
			},
		},
	})
}

func testAccUserGroupsDataSourceConfig_basic(name string) string {
	return `
resource "komodo_user_group" "test" {
  name = "` + name + `"
}

data "komodo_user_groups" "test" {
  depends_on = [komodo_user_group.test]
}
`
}

func TestUnitUserGroupsDataSource_configure(t *testing.T) {
	d := &UserGroupsDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}

// ─── Mock-server tests ────────────────────────────────────────────────────────

const mockUserGroupsDataSourceConfig = `
data "komodo_user_groups" "test" {}
`

func TestAccUserGroupsDataSource_listError(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"ListUserGroups": {http.StatusInternalServerError, `"list error"`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      mockUserGroupProviderConfig(srv.URL) + mockUserGroupsDataSourceConfig,
				ExpectError: regexp.MustCompile(`(?i)error`),
			},
		},
	})
}

func TestAccUserGroupsDataSource_readSuccess(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"ListUserGroups": {http.StatusOK, `[` + mockValidUserGroupJSON + `]`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupsDataSourceConfig,
				Check:  resource.TestCheckResourceAttr("data.komodo_user_groups.test", "groups.#", "1"),
			},
		},
	})
}

func TestAccUserGroupsDataSource_readSuccessWithAllDirect(t *testing.T) {
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"ListUserGroups": {http.StatusOK, `[` + mockValidUserGroupWithUsersJSON + `]`},
	})
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: mockUserGroupProviderConfig(srv.URL) + mockUserGroupsDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_user_groups.test", "groups.#", "1"),
					resource.TestCheckResourceAttr("data.komodo_user_groups.test", "groups.0.all.key", "val"),
				),
			},
		},
	})
}

// TestUnitUserGroupsDataSource_readConfigGetErrorDirect covers the Config.Get HasError branch.
func TestUnitUserGroupsDataSource_readConfigGetErrorDirect(t *testing.T) {
	d := &UserGroupsDataSource{}
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

// TestUnitUserGroupsDataSource_readWithNonEmptyAllDirect covers the non-empty All loop body.
func TestUnitUserGroupsDataSource_readWithNonEmptyAllDirect(t *testing.T) {
	const listWithAllJSON = `[` + mockValidUserGroupWithUsersJSON + `]`
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"ListUserGroups": {http.StatusOK, listWithAllJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	d := &UserGroupsDataSource{client: c}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)
	configVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"groups": tftypes.NewValue(tftypes.List{ElementType: schemaType.(tftypes.Object).AttributeTypes["groups"].(tftypes.List).ElementType}, nil),
	})
	req := datasource.ReadRequest{Config: tfsdk.Config{Raw: configVal, Schema: schemaResp.Schema}}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
	var stateData UserGroupsDataSourceModel
	resp.State.Get(ctx, &stateData)
	if len(stateData.Groups) == 0 {
		t.Fatal("expected non-empty groups in state")
	}
	if len(stateData.Groups[0].All.Elements()) == 0 {
		t.Fatalf("expected non-empty All in group, got: %v", stateData.Groups[0].All)
	}
}

// TestUnitUserGroupsDataSource_readNilUsersDirect covers the nil userIDs branch.
func TestUnitUserGroupsDataSource_readNilUsersDirect(t *testing.T) {
	const listWithNullUsersJSON = `[` + mockValidUserGroupNullUsersJSON + `]`
	srv := newUserGroupMockServer(t, map[string]mockUserGroupRoute{
		"ListUserGroups": {http.StatusOK, listWithNullUsersJSON},
	})
	c := client.NewClient(srv.URL, "mock-user", "mock-pass")
	d := &UserGroupsDataSource{client: c}
	ctx := context.Background()
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, datasource.SchemaRequest{}, schemaResp)
	schemaType := schemaResp.Schema.Type().TerraformType(ctx)
	configVal := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"groups": tftypes.NewValue(tftypes.List{ElementType: schemaType.(tftypes.Object).AttributeTypes["groups"].(tftypes.List).ElementType}, nil),
	})
	req := datasource.ReadRequest{Config: tfsdk.Config{Raw: configVal, Schema: schemaResp.Schema}}
	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected error: %v", resp.Diagnostics)
	}
}
