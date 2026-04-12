// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import "time"

// msToRFC3339 converts a Unix millisecond timestamp to an RFC3339 string.
// Zero is mapped to "" (meaning "no expiration" or "not set").
func msToRFC3339(ms int64) string {
	if ms == 0 {
		return ""
	}
	return time.UnixMilli(ms).UTC().Format(time.RFC3339)
}

// rfc3339ToMs converts an RFC3339 string to a Unix millisecond timestamp.
// Empty string is mapped to 0 (no expiration / not set).
func rfc3339ToMs(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}
