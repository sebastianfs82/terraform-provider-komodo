// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccUsersDataSource_basic verifies the list returns at least the admin
// user that the acceptance test suite authenticates as.
func TestAccUsersDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "komodo_users" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_users.all", "users.#"),
				),
			},
		},
	})
}

func TestAccUsersDataSource_hasFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "komodo_users" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_users.all", "users.0.id"),
					resource.TestCheckResourceAttrSet("data.komodo_users.all", "users.0.username"),
					resource.TestCheckResourceAttrSet("data.komodo_users.all", "users.0.enabled"),
				),
			},
		},
	})
}
