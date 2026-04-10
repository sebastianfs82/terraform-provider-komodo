// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccStackResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_basic("tf-test-stack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-test-stack"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
				),
			},
		},
	})
}

func TestAccStackResource_withFileContents(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_withFileContents("tf-test-stack-inline"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-test-stack-inline"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "files.contents"),
				),
			},
		},
	})
}

func TestAccStackResource_withGit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_withGit(
					"tf-test-stack-git",
					"owner/my-stack-repo",
					"main",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-test-stack-git"),
					resource.TestCheckResourceAttr("komodo_stack.test", "source.url", "https://github.com"),
					resource.TestCheckResourceAttr("komodo_stack.test", "source.path", "owner/my-stack-repo"),
					resource.TestCheckResourceAttr("komodo_stack.test", "source.branch", "main"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
				),
			},
		},
	})
}

// TestAccStackResource_gitRecloneComputed verifies that omitting reclone from a
// git block does not cause a "provider produced inconsistent result" error.
// The API always returns the zero value (false) for reclone; without Computed:true
// on the schema attribute, the planned null vs actual false mismatch was fatal.
func TestAccStackResource_gitRecloneComputed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Apply without reclone — must succeed without an inconsistency error.
			{
				Config: testAccStackResourceConfig_withGit(
					"tf-test-stack-reclone",
					"owner/my-stack-repo",
					"main",
				),
				Check: resource.TestCheckResourceAttr("komodo_stack.test", "source.reclone_enforced", "false"),
			},
			// Re-plan with the same config — must produce an empty diff.
			{
				Config:             testAccStackResourceConfig_withGit("tf-test-stack-reclone", "owner/my-stack-repo", "main"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccStackResource_withPreDeploy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_withPreDeploy("tf-test-stack-pre"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-test-stack-pre"),
					resource.TestCheckResourceAttr("komodo_stack.test", "pre_deploy.command", "echo pre"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
				),
			},
		},
	})
}

func TestAccStackResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_withGit("tf-update-stack", "owner/repo", "main"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "source.branch", "main"),
				),
			},
			{
				Config: testAccStackResourceConfig_withGit("tf-update-stack", "owner/repo", "develop"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "source.branch", "develop"),
				),
			},
		},
	})
}

func TestAccStackResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_basic("tf-import-stack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "name", "tf-import-stack"),
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
				),
			},
			{
				ResourceName:      "komodo_stack.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStackResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccStackResourceConfig_basic("tf-disappear-stack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_stack.test", "id"),
					testAccStackDisappears("komodo_stack.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccStackDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteStack(context.Background(), rs.Primary.ID)
	}
}

// Test configuration functions

func testAccStackResourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"
}
`, name)
}

func testAccStackResourceConfig_withFileContents(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  compose = {
    contents = <<-EOT
      services:
        web:
          image: nginx:latest
    EOT
  }
}
`, name)
}

func testAccStackResourceConfig_withGit(name, repo, branch string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  source = {
    url    = "https://github.com"
    path   = "%s"
    branch = "%s"
  }
}
`, name, repo, branch)
}

func testAccStackResourceConfig_withPreDeploy(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  pre_deploy = {
    command = "echo pre"
  }
}
`, name)
}

// TestAccStackResource_preDeployPathWithoutCommand verifies that setting path
// without command is rejected at plan time.
func TestAccStackResource_preDeployPathWithoutCommand(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackResourceConfig_preDeployPathOnly("tf-test-stack-pre-invalid"),
				ExpectError: regexp.MustCompile(`path requires command`),
			},
		},
	})
}

// TestAccStackResource_postDeployPathWithoutCommand verifies that setting path
// without command is rejected at plan time for post_deploy.
func TestAccStackResource_postDeployPathWithoutCommand(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccStackResourceConfig_postDeployPathOnly("tf-test-stack-post-invalid"),
				ExpectError: regexp.MustCompile(`path requires command`),
			},
		},
	})
}

// TestAccStackResource_gitRepoConflicts verifies that setting source.repo_id
// alongside any direct-clone field is rejected at plan time.
func TestAccStackResource_gitRepoConflicts(t *testing.T) {
	cases := []struct {
		name   string
		config string
	}{
		{
			name: "repo_and_url",
			config: `
resource "komodo_stack" "test" {
  name = "tf-test-conflict"
  source = {
    repo_id = "my-git-repo"
    url  = "https://github.com"
  }
}`,
		},
		{
			name: "repo_and_path",
			config: `
resource "komodo_stack" "test" {
  name = "tf-test-conflict"
  source = {
    repo_id = "my-git-repo"
    path = "owner/repo"
  }
}`,
		},
		{
			name: "repo_and_branch",
			config: `
resource "komodo_stack" "test" {
  name = "tf-test-conflict"
  source = {
    repo_id = "my-git-repo"
    branch = "main"
  }
}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						ExpectError: regexp.MustCompile(`source\.repo_id conflicts with other source fields`),
					},
				},
			})
		})
	}
}

func testAccStackResourceConfig_preDeployPathOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  pre_deploy = {
    path = "/opt/app"
  }
}
`, name)
}

func testAccStackResourceConfig_postDeployPathOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  post_deploy = {
    path = "/opt/app"
  }
}
`, name)
}

// TestAccStackResource_preDeployPartialFieldRemoval verifies that removing only
// one field (path or command) from a pre_deploy block is applied correctly —
// i.e. the plan is non-empty and the removed field is cleared on the API.
func TestAccStackResource_preDeployPartialFieldRemoval(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with both path and command set.
			{
				Config: testAccStackResourceConfig_preDeployBothFields("tf-test-stack-pre-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "pre_deploy.path", "/opt/app"),
					resource.TestCheckResourceAttr("komodo_stack.test", "pre_deploy.command", "echo pre"),
				),
			},
			// Step 2: remove path only — must produce a diff and clear path.
			{
				Config: testAccStackResourceConfig_preDeployCommandOnly("tf-test-stack-pre-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "pre_deploy.command", "echo pre"),
					resource.TestCheckNoResourceAttr("komodo_stack.test", "pre_deploy.path"),
				),
			},
			// Step 3: remove command only (path already gone) — clears the whole block.
			{
				Config: testAccStackResourceConfig_basic("tf-test-stack-pre-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_stack.test", "pre_deploy.command"),
					resource.TestCheckNoResourceAttr("komodo_stack.test", "pre_deploy.path"),
				),
			},
		},
	})
}

// TestAccStackResource_postDeployPartialFieldRemoval mirrors the pre_deploy
// test for post_deploy.
func TestAccStackResource_postDeployPartialFieldRemoval(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: create with both path and command set.
			{
				Config: testAccStackResourceConfig_postDeployBothFields("tf-test-stack-post-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "post_deploy.path", "/opt/app"),
					resource.TestCheckResourceAttr("komodo_stack.test", "post_deploy.command", "echo post"),
				),
			},
			// Step 2: remove path only — must produce a diff and clear path.
			{
				Config: testAccStackResourceConfig_postDeployCommandOnly("tf-test-stack-post-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_stack.test", "post_deploy.command", "echo post"),
					resource.TestCheckNoResourceAttr("komodo_stack.test", "post_deploy.path"),
				),
			},
			// Step 3: remove the whole post_deploy block.
			{
				Config: testAccStackResourceConfig_basic("tf-test-stack-post-partial"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("komodo_stack.test", "post_deploy.command"),
					resource.TestCheckNoResourceAttr("komodo_stack.test", "post_deploy.path"),
				),
			},
		},
	})
}

func testAccStackResourceConfig_preDeployBothFields(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  pre_deploy = {
    path    = "/opt/app"
    command = "echo pre"
  }
}
`, name)
}

func testAccStackResourceConfig_preDeployCommandOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  pre_deploy = {
    command = "echo pre"
  }
}
`, name)
}

func testAccStackResourceConfig_postDeployBothFields(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  post_deploy = {
    path    = "/opt/app"
    command = "echo post"
  }
}
`, name)
}

func testAccStackResourceConfig_postDeployCommandOnly(name string) string {
	return fmt.Sprintf(`
resource "komodo_stack" "test" {
  name = "%s"

  post_deploy = {
    command = "echo post"
  }
}
`, name)
}
