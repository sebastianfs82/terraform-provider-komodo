package client

import (
	"encoding/json"
	"fmt"
)

// Duplicate OID type found in client.go; remove this file to avoid redeclaration.
type OID struct {
	OID string `json:"$oid"`
}

func (o *OID) UnmarshalJSON(data []byte) error {
	var aux map[string]string
	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("OID: failed to unmarshal: %w", err)
	}
	if oid, ok := aux["$oid"]; ok {
		o.OID = oid
		return nil
	}
	return fmt.Errorf("OID: missing $oid field in %s", string(data))
}

func (o OID) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"$oid": o.OID})
}
