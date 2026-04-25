// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
	"time"
)

func TestTimestamps_msToRFC3339_zero(t *testing.T) {
	if got := msToRFC3339(0); got != "" {
		t.Errorf("msToRFC3339(0) = %q, want empty string", got)
	}
}

func TestTimestamps_msToRFC3339_nonZero(t *testing.T) {
	// 2024-01-15T12:00:00Z in milliseconds
	ms := int64(1705320000000)
	got := msToRFC3339(ms)
	if got == "" {
		t.Fatal("msToRFC3339 returned empty string for non-zero input")
	}
	// Round-trip: parse back and compare
	parsed, err := time.Parse(time.RFC3339, got)
	if err != nil {
		t.Fatalf("msToRFC3339 returned non-RFC3339 string %q: %v", got, err)
	}
	if parsed.UnixMilli() != ms {
		t.Errorf("round-trip failed: got %d ms, want %d ms", parsed.UnixMilli(), ms)
	}
}

func TestTimestamps_rfc3339ToMs_empty(t *testing.T) {
	ms, err := rfc3339ToMs("")
	if err != nil {
		t.Fatalf("rfc3339ToMs(\"\") returned error: %v", err)
	}
	if ms != 0 {
		t.Errorf("rfc3339ToMs(\"\") = %d, want 0", ms)
	}
}

func TestTimestamps_rfc3339ToMs_valid(t *testing.T) {
	const s = "2024-01-15T12:00:00Z"
	ms, err := rfc3339ToMs(s)
	if err != nil {
		t.Fatalf("rfc3339ToMs(%q) returned error: %v", s, err)
	}
	if ms == 0 {
		t.Error("rfc3339ToMs returned 0 for a non-empty timestamp")
	}
	// Round-trip: convert back
	got := msToRFC3339(ms)
	if got != s {
		t.Errorf("round-trip: msToRFC3339(rfc3339ToMs(%q)) = %q", s, got)
	}
}

func TestTimestamps_rfc3339ToMs_invalid(t *testing.T) {
	_, err := rfc3339ToMs("not-a-date")
	if err == nil {
		t.Error("rfc3339ToMs expected error for invalid input, got nil")
	}
}

func TestTimestamps_msToRFC3339RoundTrip(t *testing.T) {
	cases := []string{
		"2024-01-01T00:00:00Z",
		"2025-12-31T23:59:59Z",
		"2026-06-15T10:30:00Z",
	}
	for _, s := range cases {
		t.Run(s, func(t *testing.T) {
			ms, err := rfc3339ToMs(s)
			if err != nil {
				t.Fatalf("rfc3339ToMs(%q) error: %v", s, err)
			}
			got := msToRFC3339(ms)
			if got != s {
				t.Errorf("round-trip failed: got %q, want %q", got, s)
			}
		})
	}
}
