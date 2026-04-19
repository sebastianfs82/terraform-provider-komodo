// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"regexp"
	"testing"

	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUnitResourceSyncDataSource_configure(t *testing.T) {
	d := &ResourceSyncDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}

func TestAccResourceSyncDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceSyncDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccResourceSyncDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceSyncDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccResourceSyncDataSourceConfig_bothSet() string {
	return `
data "komodo_resource_sync" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccResourceSyncDataSourceConfig_neitherSet() string {
	return `
data "komodo_resource_sync" "test" {}
`
}
