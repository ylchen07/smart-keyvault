package models

// Vault represents a secrets vault/backend
type Vault struct {
	Name     string            `json:"name"`
	Provider string            `json:"provider"`
	Metadata map[string]string `json:"metadata,omitempty"`
}
