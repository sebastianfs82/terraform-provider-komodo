// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"math/big"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ─── DeploymentAction tests ───────────────────────────────────────────────────

func TestUnitDeploymentAction_metadata(t *testing.T) {
	a := NewDeploymentAction()
	req := action.MetadataRequest{ProviderTypeName: "komodo"}
	resp := &action.MetadataResponse{}
	a.Metadata(context.Background(), req, resp)
	if resp.TypeName != "komodo_deployment" {
		t.Fatalf("expected TypeName komodo_deployment, got %q", resp.TypeName)
	}
}

func TestUnitDeploymentAction_schema(t *testing.T) {
	a := &DeploymentAction{}
	resp := &action.SchemaResponse{}
	a.Schema(context.Background(), action.SchemaRequest{}, resp)
	if resp.Schema.MarkdownDescription == "" {
		t.Fatal("expected non-empty schema MarkdownDescription")
	}
	for _, attr := range []string{"id", "action", "stop_signal", "stop_time", "signal", "time"} {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Fatalf("expected %q attribute in schema", attr)
		}
	}
}

func TestUnitDeploymentAction_configure(t *testing.T) {
	ctx := context.Background()

	t.Run("nil_noop", func(t *testing.T) {
		a := &DeploymentAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: nil}, cfgResp)
		if cfgResp.Diagnostics.HasError() {
			t.Fatal("expected no error for nil ProviderData")
		}
	})

	t.Run("wrong_type", func(t *testing.T) {
		a := &DeploymentAction{}
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: "bad"}, cfgResp)
		if !cfgResp.Diagnostics.HasError() {
			t.Fatal("expected error for wrong ProviderData type")
		}
	})

	t.Run("valid_client", func(t *testing.T) {
		a := &DeploymentAction{}
		_, c := newActionSuccessMockServer(t)
		cfgResp := &action.ConfigureResponse{}
		a.Configure(ctx, action.ConfigureRequest{ProviderData: c}, cfgResp)
		if cfgResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", cfgResp.Diagnostics)
		}
		if a.client == nil {
			t.Fatal("expected client to be set")
		}
	})
}

// buildDeploymentActionInvokeReq builds a DeploymentAction InvokeRequest with all optional fields null.
func buildDeploymentActionInvokeReq(ctx context.Context, schm action.SchemaResponse, depID, act string) action.InvokeRequest {
	schemaType := schm.Schema.Type().TerraformType(ctx)
	raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
		"id":          tftypes.NewValue(tftypes.String, depID),
		"action":      tftypes.NewValue(tftypes.String, act),
		"stop_signal": tftypes.NewValue(tftypes.String, nil),
		"stop_time":   tftypes.NewValue(tftypes.Number, nil),
		"signal":      tftypes.NewValue(tftypes.String, nil),
		"time":        tftypes.NewValue(tftypes.Number, nil),
	})
	return action.InvokeRequest{Config: tfsdk.Config{Schema: schm.Schema, Raw: raw}}
}

func TestUnitDeploymentAction_invoke(t *testing.T) {
	ctx := context.Background()

	getSchema := func() action.SchemaResponse {
		a := &DeploymentAction{}
		resp := &action.SchemaResponse{}
		a.Schema(ctx, action.SchemaRequest{}, resp)
		return *resp
	}

	cases := []struct {
		name string
		act  string
	}{
		{"deploy", "deploy"},
		{"start", "start"},
		{"stop", "stop"},
		{"restart", "restart"},
		{"pull", "pull"},
		{"pause", "pause"},
		{"unpause", "unpause"},
		{"destroy", "destroy"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name+"_success", func(t *testing.T) {
			_, c := newActionSuccessMockServer(t)
			a := &DeploymentAction{client: c}
			schResp := getSchema()
			invokeResp := &action.InvokeResponse{}
			a.Invoke(ctx, buildDeploymentActionInvokeReq(ctx, schResp, "dep-id", tc.act), invokeResp)
			if invokeResp.Diagnostics.HasError() {
				t.Fatalf("unexpected error for action %q: %s", tc.act, invokeResp.Diagnostics)
			}
		})

		tc2 := tc
		t.Run(tc2.name+"_client_error", func(t *testing.T) {
			_, c := newActionErrorMockServer(t)
			a := &DeploymentAction{client: c}
			schResp := getSchema()
			invokeResp := &action.InvokeResponse{}
			a.Invoke(ctx, buildDeploymentActionInvokeReq(ctx, schResp, "dep-id", tc2.act), invokeResp)
			if !invokeResp.Diagnostics.HasError() {
				t.Fatalf("expected error for action %q on client failure", tc2.act)
			}
		})
	}

	t.Run("deploy_with_stop_signal", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &DeploymentAction{client: c}
		schResp := getSchema()
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":          tftypes.NewValue(tftypes.String, "dep-id"),
			"action":      tftypes.NewValue(tftypes.String, "deploy"),
			"stop_signal": tftypes.NewValue(tftypes.String, "SIGKILL"),
			"stop_time":   tftypes.NewValue(tftypes.Number, nil),
			"signal":      tftypes.NewValue(tftypes.String, nil),
			"time":        tftypes.NewValue(tftypes.Number, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("stop_with_signal", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &DeploymentAction{client: c}
		schResp := getSchema()
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":          tftypes.NewValue(tftypes.String, "dep-id"),
			"action":      tftypes.NewValue(tftypes.String, "stop"),
			"stop_signal": tftypes.NewValue(tftypes.String, nil),
			"stop_time":   tftypes.NewValue(tftypes.Number, nil),
			"signal":      tftypes.NewValue(tftypes.String, "SIGTERM"),
			"time":        tftypes.NewValue(tftypes.Number, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("invalid_action", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &DeploymentAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildDeploymentActionInvokeReq(ctx, schResp, "dep-id", "bogus"), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid action")
		}
	})

	t.Run("config_error", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &DeploymentAction{client: c}
		schResp := getSchema()
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, buildActionConfigErrorReq(ctx, schResp.Schema), invokeResp)
		if !invokeResp.Diagnostics.HasError() {
			t.Fatal("expected error for invalid config")
		}
	})

	t.Run("deploy_with_stop_time", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &DeploymentAction{client: c}
		schResp := getSchema()
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":          tftypes.NewValue(tftypes.String, "dep-id"),
			"action":      tftypes.NewValue(tftypes.String, "deploy"),
			"stop_signal": tftypes.NewValue(tftypes.String, nil),
			"stop_time":   tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(30)),
			"signal":      tftypes.NewValue(tftypes.String, nil),
			"time":        tftypes.NewValue(tftypes.Number, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("stop_with_time", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &DeploymentAction{client: c}
		schResp := getSchema()
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":          tftypes.NewValue(tftypes.String, "dep-id"),
			"action":      tftypes.NewValue(tftypes.String, "stop"),
			"stop_signal": tftypes.NewValue(tftypes.String, nil),
			"stop_time":   tftypes.NewValue(tftypes.Number, nil),
			"signal":      tftypes.NewValue(tftypes.String, nil),
			"time":        tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(30)),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("destroy_with_signal", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &DeploymentAction{client: c}
		schResp := getSchema()
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":          tftypes.NewValue(tftypes.String, "dep-id"),
			"action":      tftypes.NewValue(tftypes.String, "destroy"),
			"stop_signal": tftypes.NewValue(tftypes.String, nil),
			"stop_time":   tftypes.NewValue(tftypes.Number, nil),
			"signal":      tftypes.NewValue(tftypes.String, "SIGTERM"),
			"time":        tftypes.NewValue(tftypes.Number, nil),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})

	t.Run("destroy_with_time", func(t *testing.T) {
		_, c := newActionSuccessMockServer(t)
		a := &DeploymentAction{client: c}
		schResp := getSchema()
		schemaType := schResp.Schema.Type().TerraformType(ctx)
		raw := tftypes.NewValue(schemaType, map[string]tftypes.Value{
			"id":          tftypes.NewValue(tftypes.String, "dep-id"),
			"action":      tftypes.NewValue(tftypes.String, "destroy"),
			"stop_signal": tftypes.NewValue(tftypes.String, nil),
			"stop_time":   tftypes.NewValue(tftypes.Number, nil),
			"signal":      tftypes.NewValue(tftypes.String, nil),
			"time":        tftypes.NewValue(tftypes.Number, new(big.Float).SetInt64(30)),
		})
		invokeResp := &action.InvokeResponse{}
		a.Invoke(ctx, action.InvokeRequest{Config: tfsdk.Config{Schema: schResp.Schema, Raw: raw}}, invokeResp)
		if invokeResp.Diagnostics.HasError() {
			t.Fatalf("unexpected error: %s", invokeResp.Diagnostics)
		}
	})
}
