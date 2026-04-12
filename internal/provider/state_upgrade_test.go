// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// ---------------------------------------------------------------------------
// Current-schema attribute-name tests
// ---------------------------------------------------------------------------

// TestActionSchema_currentFieldNames verifies the current action schema
// uses the new attribute names and does not retain the old ones.
func TestActionSchema_currentFieldNames(t *testing.T) {
	ctx := context.Background()
	r := &ActionResource{}
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)

	attrs := resp.Schema.Attributes

	for _, newName := range []string{"run_on_startup_enabled", "reload_dependencies_enabled"} {
		if _, ok := attrs[newName]; !ok {
			t.Errorf("current action schema missing expected attribute %q", newName)
		}
	}
	for _, oldName := range []string{"run_at_startup", "reload_deno_deps"} {
		if _, ok := attrs[oldName]; ok {
			t.Errorf("current action schema should not contain deprecated attribute %q", oldName)
		}
	}
}

// TestActionSchema_scheduleAlertEnabled verifies the schedule block in the
// current action schema has alert_enabled (not the old alert attribute).
func TestActionSchema_scheduleAlertEnabled(t *testing.T) {
	ctx := context.Background()
	r := &ActionResource{}
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)

	scheduleAttr, ok := resp.Schema.Attributes["schedule"]
	if !ok {
		t.Fatal("current action schema missing schedule attribute")
	}
	nested, ok := scheduleAttr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("schedule is not SingleNestedAttribute")
	}
	if _, ok := nested.Attributes["alert_enabled"]; !ok {
		t.Error("current action schema schedule missing alert_enabled")
	}
	if _, ok := nested.Attributes["alert"]; ok {
		t.Error("current action schema schedule should not contain deprecated alert")
	}
}

// TestProcedureSchema_scheduleAlertEnabled verifies the procedure schedule
// block has alert_enabled, not the old alert attribute.
func TestProcedureSchema_scheduleAlertEnabled(t *testing.T) {
	ctx := context.Background()
	r := &ProcedureResource{}
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)

	scheduleAttr, ok := resp.Schema.Attributes["schedule"]
	if !ok {
		t.Fatal("current procedure schema missing schedule attribute")
	}
	nested, ok := scheduleAttr.(schema.SingleNestedAttribute)
	if !ok {
		t.Fatal("schedule is not SingleNestedAttribute")
	}
	if _, ok := nested.Attributes["alert_enabled"]; !ok {
		t.Error("current procedure schema schedule missing alert_enabled")
	}
	if _, ok := nested.Attributes["alert"]; ok {
		t.Error("current procedure schema schedule should not contain deprecated alert")
	}
}
