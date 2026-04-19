// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVersionDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVersionDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_version.test", "version"),
					resource.TestCheckResourceAttrSet("data.komodo_version.test", "id"),
					resource.TestCheckResourceAttrPair("data.komodo_version.test", "id", "data.komodo_version.test", "version"),
				),
			},
		},
	})
}

const testAccVersionDataSourceConfig = `
data "komodo_version" "test" {}
`

func TestUnitVersionDataSource_configure(t *testing.T) {
	d := &VersionDataSource{}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected diagnostic error for wrong provider data type")
	}
}
