package client

// OnboardingKey represents an onboarding key entity.
type OnboardingKey struct {
	PublicKey     string   `json:"public_key"`
	Enabled       bool     `json:"enabled"`
	Expires       int64    `json:"expires"`
	Name          string   `json:"name"`
	Onboarded     []string `json:"onboarded"`
	CreatedAt     int64    `json:"created_at"`
	Tags          []string `json:"tags"`
	Privileged    bool     `json:"privileged"`
	CopyServer    string   `json:"copy_server"`
	CreateBuilder bool     `json:"create_builder"`
}

// CreateOnboardingKeyRequest holds the parameters for creating an onboarding key.
type CreateOnboardingKeyRequest struct {
	Name          string   `json:"name"`
	Expires       int64    `json:"expires"`
	PrivateKey    *string  `json:"private_key,omitempty"`
	Tags          []string `json:"tags,omitempty"`
	Privileged    bool     `json:"privileged"`
	CopyServer    string   `json:"copy_server"`
	CreateBuilder bool     `json:"create_builder"`
}

// CreateOnboardingKeyResponse is the response from creating an onboarding key.
type CreateOnboardingKeyResponse struct {
	PrivateKey string        `json:"private_key"`
	Created    OnboardingKey `json:"created"`
}

// UpdateOnboardingKeyRequest holds the parameters for updating an onboarding key.
type UpdateOnboardingKeyRequest struct {
	PublicKey     string    `json:"public_key"`
	Enabled       *bool     `json:"enabled,omitempty"`
	Name          *string   `json:"name,omitempty"`
	Expires       *int64    `json:"expires,omitempty"`
	Tags          *[]string `json:"tags,omitempty"`
	Privileged    *bool     `json:"privileged,omitempty"`
	CopyServer    *string   `json:"copy_server,omitempty"`
	CreateBuilder *bool     `json:"create_builder,omitempty"`
}

// DeleteOnboardingKeyRequest holds the parameters for deleting an onboarding key.
type DeleteOnboardingKeyRequest struct {
	PublicKey string `json:"public_key"`
}
