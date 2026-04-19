// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"context"
	"regexp"
	"testing"
	datasource "github.com/hashicorp/terraform-plugin-framework/datasource"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProcedureDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureDataSourceConfig_basic("tf-acc-procedure-ds-basic"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.komodo_procedure.test", "id"),
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "name", "tf-acc-procedure-ds-basic"),
				),
			},
		},
	})
}

func TestAccProcedureDataSource_fields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureDataSourceConfig_withSchedule("tf-acc-procedure-ds-fields"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "name", "tf-acc-procedure-ds-fields"),
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "schedule.format", "Cron"),
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "schedule.expression", "0 * * * *"),
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "schedule.enabled", "true"),
					resource.TestCheckResourceAttrSet("data.komodo_procedure.test", "failure_alert_enabled"),
					resource.TestCheckResourceAttrSet("data.komodo_procedure.test", "webhook.enabled"),
				),
			},
		},
	})
}

func testAccProcedureDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "src" {
  name = %q
}

data "komodo_procedure" "test" {
  id         = komodo_procedure.src.id
  depends_on = [komodo_procedure.src]
}
`, name)
}

func testAccProcedureDataSourceConfig_withSchedule(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "src" {
  name = %q
  schedule {
    format     = "Cron"
    expression = "0 * * * *"
    enabled    = true
  }
}

data "komodo_procedure" "test" {
  id         = komodo_procedure.src.id
  depends_on = [komodo_procedure.src]
}
`, name)
}

func TestAccProcedureDataSource_stages(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureDataSourceConfig_withStages("tf-acc-procedure-ds-stages"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "name", "tf-acc-procedure-ds-stages"),
					// data source exposes stages as a JSON string
					resource.TestCheckResourceAttrSet("data.komodo_procedure.test", "stages"),
				),
			},
		},
	})
}

func TestAccProcedureDataSource_failureAlertEnabled(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProcedureDataSourceConfig_withFailureAlert("tf-acc-procedure-ds-failure-alert"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.komodo_procedure.test", "failure_alert_enabled", "false"),
				),
			},
		},
	})
}

func testAccProcedureDataSourceConfig_withStages(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "child" {
  name = "%s-child"
}

resource "komodo_procedure" "src" {
  name = %q

  stage {
    name = "Run"

    execution {
      type = "RunProcedure"
      parameters = {
        procedure = komodo_procedure.child.id
      }
    }
  }
}

data "komodo_procedure" "test" {
  id         = komodo_procedure.src.id
  depends_on = [komodo_procedure.src]
}
`, name, name)
}

func testAccProcedureDataSourceConfig_withFailureAlert(name string) string {
	return fmt.Sprintf(`
resource "komodo_procedure" "src" {
  name                  = %q
  failure_alert_enabled = false
}

data "komodo_procedure" "test" {
  id         = komodo_procedure.src.id
  depends_on = [komodo_procedure.src]
}
`, name)
}

func TestUnitProcedureDataSource_configure(t *testing.T) {
d := &ProcedureDataSource{}
resp := &datasource.ConfigureResponse{}
d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "wrong"}, resp)
if !resp.Diagnostics.HasError() {
t.Fatal("expected diagnostic error for wrong provider data type")
}
}

func TestAccProcedureDataSource_bothSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProcedureDataSourceConfig_bothSet(),
				ExpectError: regexp.MustCompile(`Only one of`),
			},
		},
	})
}

func TestAccProcedureDataSource_neitherSet_isError(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccProcedureDataSourceConfig_neitherSet(),
				ExpectError: regexp.MustCompile(`One of`),
			},
		},
	})
}

func testAccProcedureDataSourceConfig_bothSet() string {
	return `
data "komodo_procedure" "test" {
  id   = "some-id"
  name = "some-name"
}
`
}

func testAccProcedureDataSourceConfig_neitherSet() string {
	return `
data "komodo_procedure" "test" {}
`
}
