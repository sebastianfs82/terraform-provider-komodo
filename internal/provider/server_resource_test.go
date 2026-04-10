// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sebastianfs82/terraform-provider-komodo/internal/client"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccServerResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfig("tf-acc-server-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-basic"),
					resource.TestCheckResourceAttrSet("komodo_server.test", "id"),
				),
			},
		},
	})
}

func TestAccServerResource_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfigWithRegion("tf-acc-server-update", "us-east"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-update"),
					resource.TestCheckResourceAttr("komodo_server.test", "region", "us-east"),
				),
			},
			{
				Config: testAccServerResourceConfigWithRegion("tf-acc-server-update", "us-west"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-update"),
					resource.TestCheckResourceAttr("komodo_server.test", "region", "us-west"),
				),
			},
		},
	})
}

func TestAccServerResource_importState(t *testing.T) {
	var serverID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfig("tf-acc-server-import"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_server.test", "id"),
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["komodo_server.test"]
						if !ok {
							return fmt.Errorf("resource not found")
						}
						serverID = rs.Primary.ID
						return nil
					},
				),
			},
			{
				Config:            testAccServerResourceConfig("tf-acc-server-import"),
				ResourceName:      "komodo_server.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(_ *terraform.State) (string, error) { return serverID, nil },
			},
		},
	})
}

func TestAccServerResource_disappears(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfig("tf-acc-server-disappears"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("komodo_server.test", "id"),
					testAccServerDisappears("komodo_server.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServerResource_withAlertThresholds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfigWithThresholds("tf-acc-server-thresholds", 80.0, 95.0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-thresholds"),
					resource.TestCheckResourceAttr("komodo_server.test", "cpu_warning", "80"),
					resource.TestCheckResourceAttr("komodo_server.test", "cpu_critical", "95"),
				),
			},
		},
	})
}

func TestAccServerResource_withLinks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerResourceConfigWithLinks("tf-acc-server-links", []string{"http://grafana.local", "http://prometheus.local"}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("komodo_server.test", "name", "tf-acc-server-links"),
					resource.TestCheckResourceAttr("komodo_server.test", "links.#", "2"),
					resource.TestCheckResourceAttr("komodo_server.test", "links.0", "http://grafana.local"),
					resource.TestCheckResourceAttr("komodo_server.test", "links.1", "http://prometheus.local"),
				),
			},
		},
	})
}

func testAccServerDisappears(resourceName string) resource.TestCheckFunc {
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
		return c.DeleteServer(context.Background(), rs.Primary.ID)
	}
}

func testAccServerResourceConfig(name string) string {
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name = %q
}
`, name)
}

func testAccServerResourceConfigWithRegion(name, region string) string {
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name   = %q
  region = %q
}
`, name, region)
}

func testAccServerResourceConfigWithThresholds(name string, cpuWarn, cpuCrit float64) string {
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name         = %q
  cpu_warning  = %g
  cpu_critical = %g
}
`, name, cpuWarn, cpuCrit)
}

func testAccServerResourceConfigWithLinks(name string, links []string) string {
	linksTF := "["
	for i, l := range links {
		if i > 0 {
			linksTF += ", "
		}
		linksTF += fmt.Sprintf("%q", l)
	}
	linksTF += "]"
	return fmt.Sprintf(`
resource "komodo_server" "test" {
  name  = %q
  links = %s
}
`, name, linksTF)
}
