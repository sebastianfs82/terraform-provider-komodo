// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"fmt"
)

// Builder represents a Komodo builder resource (Resource<BuilderConfig>).
type Builder struct {
	ID          OID           `json:"_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Config      BuilderConfig `json:"config"`
}

// BuilderConfig is a discriminated union of builder configuration types,
// serialised by the Komodo API as {"type": "<Variant>", "params": {...}}.
type BuilderConfig struct {
	Type   string          `json:"type"`
	Params json.RawMessage `json:"params"`
}

// GetUrlConfig decodes the Params field as UrlBuilderConfig.
// Returns nil, nil when the Type is not "Url".
func (c *BuilderConfig) GetUrlConfig() (*UrlBuilderConfig, error) {
	if c.Type != "Url" {
		return nil, nil
	}
	var cfg UrlBuilderConfig
	if err := json.Unmarshal(c.Params, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode URL builder config: %w", err)
	}
	return &cfg, nil
}

// GetServerConfig decodes the Params field as ServerBuilderConfig.
// Returns nil, nil when the Type is not "Server".
func (c *BuilderConfig) GetServerConfig() (*ServerBuilderConfig, error) {
	if c.Type != "Server" {
		return nil, nil
	}
	var cfg ServerBuilderConfig
	if err := json.Unmarshal(c.Params, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode Server builder config: %w", err)
	}
	return &cfg, nil
}

// GetAwsConfig decodes the Params field as AwsBuilderConfig.
// Returns nil, nil when the Type is not "Aws".
func (c *BuilderConfig) GetAwsConfig() (*AwsBuilderConfig, error) {
	if c.Type != "Aws" {
		return nil, nil
	}
	var cfg AwsBuilderConfig
	if err := json.Unmarshal(c.Params, &cfg); err != nil {
		return nil, fmt.Errorf("failed to decode AWS builder config: %w", err)
	}
	return &cfg, nil
}

// UrlBuilderConfig is the configuration for a Komodo URL Builder
// (connecting to an existing Periphery agent via URL).
type UrlBuilderConfig struct {
	Address            string `json:"address"`
	PeripheryPublicKey string `json:"periphery_public_key"`
	InsecureTls        bool   `json:"insecure_tls"`
	Passkey            string `json:"passkey"`
}

// ServerBuilderConfig is the configuration for a Komodo Server Builder
// (using a connected Komodo server as the builder host).
type ServerBuilderConfig struct {
	ServerID string `json:"server_id"`
}

// AwsBuilderConfig is the configuration for a Komodo AWS Builder
// (using EC2 instances spawned on-demand for each build).
type AwsBuilderConfig struct {
	Region             string   `json:"region"`
	InstanceType       string   `json:"instance_type"`
	VolumeGb           int64    `json:"volume_gb"`
	AmiID              string   `json:"ami_id"`
	SubnetID           string   `json:"subnet_id"`
	KeyPairName        string   `json:"key_pair_name"`
	AssignPublicIP     bool     `json:"assign_public_ip"`
	UsePublicIP        bool     `json:"use_public_ip"`
	SecurityGroupIDs   []string `json:"security_group_ids"`
	UserData           string   `json:"user_data"`
	Port               int64    `json:"port"`
	UseHttps           bool     `json:"use_https"`
	PeripheryPublicKey string   `json:"periphery_public_key"`
	InsecureTls        bool     `json:"insecure_tls"`
	Secrets            []string `json:"secrets"`
}

// BuilderConfigInput is a discriminated union used for write operations
// (CreateBuilder / UpdateBuilder).
type BuilderConfigInput struct {
	Type   string      `json:"type"`
	Params interface{} `json:"params"`
}

// CreateBuilderRequest is the params for the CreateBuilder Komodo API call.
type CreateBuilderRequest struct {
	Name   string             `json:"name"`
	Config BuilderConfigInput `json:"config"`
}

// UpdateBuilderRequest is the params for the UpdateBuilder Komodo API call.
type UpdateBuilderRequest struct {
	ID     string             `json:"id"`
	Config BuilderConfigInput `json:"config"`
}

// RenameBuilderRequest is the payload for the RenameBuilder write API.
type RenameBuilderRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
