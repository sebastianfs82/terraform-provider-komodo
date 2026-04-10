// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRepoDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoDataSourceConfig_basic("tf-acc-repo-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_repo.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_repo.test", "name", "tf-acc-repo-ds-basic"),
				),
			},
		},
	})
}

func TestAccRepoDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRepoDataSourceConfig_withSource("tf-acc-repo-ds-fields"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_repo.test", "name", "tf-acc-repo-ds-fields"),
					resource.TestCheckResourceAttr("data.komodo_repo.test", "source.domain", "github.com"),
					resource.TestCheckResourceAttr("data.komodo_repo.test", "source.https_enabled", "true"),
					resource.TestCheckResourceAttr("data.komodo_repo.test", "source.path", "owner/repo"),
					resource.TestCheckResourceAttr("data.komodo_repo.test", "source.branch", "main"),
					resource.TestCheckResourceAttrSet("data.komodo_repo.test", "id"),
				),
			},
		},
	})
}

func TestAccReposDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReposDataSourceConfig_basic("tf-acc-repos-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_repos.all", "repositories.#"),
				),
			},
		},
	})
}

func TestAccReposDataSource_containsCreatedRepo(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReposDataSourceConfig_basic("tf-acc-repos-ds-contains"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_repos.all", "repositories.#"),
					resource.TestCheckResourceAttrSet("data.komodo_repos.all", "repositories.0.id"),
					resource.TestCheckResourceAttrSet("data.komodo_repos.all", "repositories.0.name"),
				),
			},
		},
	})
}

func testAccRepoDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "src" {
  name = %q
}

data "komodo_repo" "test" {
  name       = komodo_repo.src.name
  depends_on = [komodo_repo.src]
}
`, name)
}

func testAccRepoDataSourceConfig_withSource(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "src" {
  name = %q
  source = {
    domain        = "github.com"
    https_enabled = true
    path          = "owner/repo"
    branch        = "main"
  }
}

data "komodo_repo" "test" {
  name       = komodo_repo.src.name
  depends_on = [komodo_repo.src]
}
`, name)
}

func testAccReposDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_repo" "src" {
  name = %q
}

data "komodo_repos" "all" {
  depends_on = [komodo_repo.src]
}
`, name)
}
