package provider

import (
	"testing"

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
