// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TrimmedStringType is a custom Terraform attribute type whose values are
// semantically equal after stripping trailing newline / carriage-return
// characters and normalising CRLF → LF.  This prevents perpetual plan drift
// when the Komodo API silently drops trailing newlines from stored strings.
type TrimmedStringType struct {
	basetypes.StringType
}

var _ basetypes.StringTypable = TrimmedStringType{}

func (t TrimmedStringType) Equal(o attr.Type) bool {
	other, ok := o.(TrimmedStringType)
	if !ok {
		return false
	}
	return t.StringType.Equal(other.StringType)
}

func (t TrimmedStringType) String() string {
	return "TrimmedStringType"
}

// ValueFromString wraps a plain StringValue in a TrimmedStringValue.
func (t TrimmedStringType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return TrimmedStringValue{StringValue: in}, diag.Diagnostics{}
}

// ValueFromTerraform converts a raw tftypes.Value into a TrimmedStringValue.
func (t TrimmedStringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	sv, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type %T", attrValue)
	}
	return TrimmedStringValue{StringValue: sv}, nil
}

// ValueType returns the Go type used for this attribute in model structs.
func (t TrimmedStringType) ValueType(_ context.Context) attr.Value {
	return TrimmedStringValue{}
}

// TrimmedStringValue is the value type for TrimmedStringType.  It implements
// StringValuableWithSemanticEquals so that Terraform considers two values equal
// when they differ only in trailing newlines / carriage returns.
type TrimmedStringValue struct {
	basetypes.StringValue
}

var _ basetypes.StringValuableWithSemanticEquals = TrimmedStringValue{}

func (v TrimmedStringValue) Type(_ context.Context) attr.Type {
	return TrimmedStringType{}
}

func (v TrimmedStringValue) Equal(o attr.Value) bool {
	other, ok := o.(TrimmedStringValue)
	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals returns true when the two values are equal after
// normalising CRLF → LF and trimming trailing newline / carriage-return chars.
func (v TrimmedStringValue) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	newValue, ok := newValuable.(TrimmedStringValue)
	if !ok {
		return false, nil
	}
	normalize := func(s string) string {
		return strings.TrimRight(strings.ReplaceAll(s, "\r\n", "\n"), "\n\r")
	}
	return normalize(v.ValueString()) == normalize(newValue.ValueString()), nil
}

// NewTrimmedStringValue constructs a known TrimmedStringValue from a Go string.
func NewTrimmedStringValue(s string) TrimmedStringValue {
	return TrimmedStringValue{StringValue: basetypes.NewStringValue(s)}
}

// NewTrimmedStringNull constructs a null TrimmedStringValue.
func NewTrimmedStringNull() TrimmedStringValue {
	return TrimmedStringValue{StringValue: basetypes.NewStringNull()}
}

// NewTrimmedStringUnknown constructs an unknown TrimmedStringValue.
func NewTrimmedStringUnknown() TrimmedStringValue {
	return TrimmedStringValue{StringValue: basetypes.NewStringUnknown()}
}
